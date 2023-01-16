package service

import (
	"common/graceful"
	"common/hashslot"
	"common/logs"
	"common/pb"
	"common/util"
	"context"
	"errors"
	"fmt"
	"metaserver/config"
	"metaserver/internal/entity"
	"metaserver/internal/usecase"
	"metaserver/internal/usecase/db"
	"metaserver/internal/usecase/logic"
	"metaserver/internal/usecase/pool"
	"strings"
	"time"

	"google.golang.org/grpc"
)

var (
	hsLog = logs.New("hash-slot-migration")
)

type HashSlotService struct {
	Store         *db.HashSlotDB
	Service       usecase.IMetadataService
	BucketService usecase.BucketService
	Cfg           *config.HashSlotConfig
	startReceive  func()
}

func NewHashSlotService(st *db.HashSlotDB, serv usecase.IMetadataService, bucketService usecase.BucketService, cfg *config.HashSlotConfig) *HashSlotService {
	return &HashSlotService{
		Store:         st,
		Service:       serv,
		BucketService: bucketService,
		Cfg:           cfg,
		startReceive:  func() {},
	}
}

func (h *HashSlotService) OnLeaderChanged(isLeader bool) {
	if isLeader {
		var (
			info  *hashslot.SlotInfo
			exist bool
			err   error
		)
		// if not exist, read from configuration
		if info, exist, err = h.Store.Get(h.Cfg.StoreID); !exist {
			if err != nil {
				logs.Std().Errorf("get slot-info when leader changed: %s", err)
				return
			}
			info = &hashslot.SlotInfo{Slots: h.Cfg.Slots, GroupID: h.Cfg.StoreID}
			logs.Std().Infof("no exist slots, init from config: id=%s, slots=%s", info.GroupID, info.Slots)
		}
		if err := logic.NewHashSlot().SaveToEtcd(h.Cfg.StoreID, info); err != nil {
			logs.Std().Error(err)
			return
		}
	}
}

func (h *HashSlotService) GetCurrentSlots(reload bool) (map[string][]string, error) {
	prov, err := h.Store.GetEdgeProvider(reload)
	if err != nil {
		return nil, err
	}
	res := make(map[string][]string)
	for _, v := range hashslot.CopyOfEdges("", prov) {
		res[v.Value] = append(res[v.Value], fmt.Sprint(v.Start, "-", v.End))
	}
	return res, nil
}

func (h *HashSlotService) PrepareMigrationTo(loc *pb.LocationInfo, slots []string) error {
	// validate slots
	provider, err := h.Store.GetEdgeProvider(false)
	if err != nil {
		return err
	}
	edges, err := hashslot.WrapSlotsToEdges(slots, pool.HttpHostPort)
	if err != nil {
		return err
	}
	// ensure all slots is in this server
	for _, edge := range edges {
		if !hashslot.IsValidEdge(edge, provider) {
			return fmt.Errorf("slot %s is not in this server", edge)
		}
	}
	// send prepare rpc to target
	cc, err := grpc.Dial(fmt.Sprint(loc.GetHost(), ":", loc.GetRpcPort()), grpc.WithInsecure())
	if err != nil {
		return err
	}
	client := pb.NewHashSlotClient(cc)
	resp, err := client.PrepareMigration(context.Background(), &pb.PrepareReq{
		Id: h.Cfg.StoreID,
		Location: &pb.LocationInfo{
			Host:     util.GetHost(),
			RpcPort:  pool.Config.Cluster.Port,
			HttpPort: pool.Config.Port,
		},
		Slots: slots,
	})
	if err != nil {
		return err
	}
	if !resp.GetSuccess() {
		return errors.New(resp.GetMessage())
	}
	// change status to migrate-to
	return h.Store.ReadyMigrateTo(loc.GetHost(), slots)
}

// PrepareMigrationFrom Change into migrate-from. Status will change back if timeout
func (h *HashSlotService) PrepareMigrationFrom(loc *pb.LocationInfo, slots []string) error {
	// validate slots
	provider, err := h.Store.GetEdgeProvider(false)
	if err != nil {
		return err
	}
	edges, err := hashslot.WrapSlotsToEdges(slots, pool.HttpHostPort)
	if err != nil {
		return err
	}
	// ensure all slots is not in this server
	for _, edge := range edges {
		if hashslot.IsValidEdge(edge, provider) {
			return fmt.Errorf("slot %s is currently in this server", edge)
		}
	}
	// change status to migrate-from
	if err := h.Store.ReadyMigrateFrom(loc.GetHost(), slots); err != nil {
		return err
	}
	cancelCh := make(chan struct{})
	h.startReceive = func() {
		close(cancelCh)
		h.startReceive = func() {}
	}
	go func() {
		select {
		case <-cancelCh:
			logs.Std().Debug("migration-from timeout-ctx canceled")
		case <-time.NewTicker(h.Cfg.PrepareTimeout).C:
			h.startReceive()
			_ = h.Store.FinishMigrateFrom()
			logs.Std().Errorf("timeout migrating from %s", loc.GetHost())
		}
	}()
	return nil
}

func (h *HashSlotService) FinishReceiveItem(success bool) error {
	h.startReceive()
	var (
		newSlots []string
		fromHost string
		ok       bool
	)
	if ok, fromHost, newSlots = h.Store.GetMigrateFrom(); !ok {
		return fmt.Errorf("get received slots fails: server is not in migrate-from")
	}
	if err := h.Store.FinishMigrateFrom(); err != nil {
		util.LogErr(err)
	}
	if !success {
		return nil
	}
	info, _, err := h.Store.Get(h.Cfg.StoreID)
	if err != nil {
		return fmt.Errorf("update slot fails after finish migration: %w", err)
	}
	// combine to new slots, ignore error because both have been validated
	curEdges, _ := hashslot.WrapSlotsToEdges(info.Slots, info.Location)
	newEdges, _ := hashslot.WrapSlotsToEdges(newSlots, info.Location)
	info.Slots = hashslot.CombineEdges(curEdges, newEdges).Strings()
	// save new slot-info
	if err = logic.NewHashSlot().SaveToEtcd(h.Cfg.StoreID, info); err != nil {
		return fmt.Errorf("save new slot-info fails after finsih migrateion: %w", err)
	}
	logs.Std().Debugf("finish migration from %s success", fromHost)
	return nil
}

func (h *HashSlotService) ReceiveItem(item *pb.MigrationItem) error {
	h.startReceive()
	var err error
	switch entity.Dest(item.Dest) {
	case entity.DestVersion:
		var i entity.Version
		if err = util.DecodeMsgp(&i, item.Data); err != nil {
			return err
		}
		err = h.Service.ReceiveVersion(item.Name, &i)
	case entity.DestMetadata:
		var i entity.Metadata
		if err = util.DecodeMsgp(&i, item.Data); err != nil {
			return err
		}
		err = h.Service.AddMetadata(item.Name, &i)
	case entity.DestBucket:
		var i entity.Bucket
		if err = util.DecodeMsgp(&i, item.Data); err != nil {
			return err
		}
		err = h.BucketService.Create(&i)
	}
	if err != nil {
		if errors.Is(err, usecase.ErrExists) {
			return nil
		}
		return err
	}
	return nil
}

// AutoMigrate migrate data
//TODO(perf): multi goroutine
func (h *HashSlotService) AutoMigrate(toLoc *pb.LocationInfo, slots []string) error {
	if ok, host, _ := h.Store.GetMigrateTo(); !ok || host != toLoc.GetHost() {
		return fmt.Errorf("no ready to migrate to %s", toLoc.GetHost())
	}
	// connect to target
	cc, err := grpc.Dial(fmt.Sprint(toLoc.Host, ":", toLoc.RpcPort), grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer cc.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stream, err := pb.NewHashSlotClient(cc).StreamingReceive(ctx)
	if err != nil {
		return err
	}
	delEdges, _ := hashslot.WrapSlotsToEdges(slots, "")
	var errs []error
	errs = append(errs, h.migrateMetadata(stream, delEdges)...)
	errs = append(errs, h.migrateBuckets(stream, delEdges)...)
	if err = h.Store.FinishMigrateTo(); err != nil {
		errs = append(errs, err)
		hsLog.Debugf("switch status to normal err: %s", err)
	}
	if len(errs) > 0 {
		sb := strings.Builder{}
		sb.WriteString("occurred errors:")
		for _, err := range errs {
			sb.WriteRune('\n')
			sb.WriteString(err.Error())
		}
		hsLog.Error(sb.String())
		return errors.New("migrate partly fails, please retry again")
	}
	// all migrate success
	// remove slots from current slot-info
	info, _, err := h.Store.Get(h.Cfg.StoreID)
	if err != nil {
		return fmt.Errorf("update slot fails after finish migration: %w", err)
	}
	curEdges, _ := hashslot.WrapSlotsToEdges(info.Slots, info.Location)
	info.Slots = hashslot.RemoveEdges(curEdges, delEdges).Strings()
	// save new slot-info
	if err = logic.NewHashSlot().SaveToEtcd(h.Cfg.StoreID, info); err != nil {
		return fmt.Errorf("save new slot-info fails after finsih migrateion: %w", err)
	}
	// close stream as success
	if err := stream.CloseSend(); err != nil {
		hsLog.Error(err)
	}
	hsLog.Infof("finish migration to %s success", toLoc.Host)
	return nil
}

func (h *HashSlotService) migrateMetadata(stream pb.HashSlot_StreamingReceiveClient, edges hashslot.EdgeList) (errs []error) {
	var sucNum int
	metaKeys := h.Service.FilterKeys(func(s string) bool {
		return hashslot.IsSlotInEdges(hashslot.CalcBytesSlot(util.StrToBytes(s)), edges)
	})
	total := len(metaKeys)
	// send all metadata and versions
	for len(metaKeys) > 0 {
		key := metaKeys[0]
		metaKeys = metaKeys[1:] // for earlier GC
		data, err := h.Service.GetMetadataBytes(key)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		// send data
		if err := stream.Send(&pb.MigrationItem{
			Name: key,
			Data: data,
			Dest: int32(entity.DestMetadata),
		}); err != nil {
			hsLog.Debugf("send metadata %s err: %s", key, err)
			errs = append(errs, err)
			continue
		}
		// recv response
		resp, err := stream.Recv()
		if err != nil {
			hsLog.Debugf("recv send-metadata %s response err: %s", key, err)
			errs = append(errs, err)
			continue
		} else if !resp.Success {
			hsLog.Debugf("send-metadata %s recv failure resposne: %s", key, resp.Message)
			errs = append(errs, errors.New(resp.Message))
			continue
		}
		// start send versions
		allVersionSuccess := true
		h.Service.ForeachVersionBytes(key, func(b []byte) bool {
			// send version
			if err := stream.Send(&pb.MigrationItem{
				Name: key,
				Data: b,
				Dest: int32(entity.DestVersion),
			}); err != nil {
				errs = append(errs, err)
				allVersionSuccess = false
				hsLog.Debugf("send-metadata-version %s err: %s", key, err)
			}
			// recv response
			resp, err := stream.Recv()
			if err != nil {
				errs = append(errs, err)
				allVersionSuccess = false
				hsLog.Debugf("send-metadata-version %s recv err: %s", key, err)
			} else if !resp.Success {
				errs = append(errs, errors.New(resp.Message))
				allVersionSuccess = false
				hsLog.Debugf("send-metadata-version %s recv failure resposne: %s", key, resp.Message)
			}
			return true
		})
		if !allVersionSuccess {
			continue
		}
		sucNum++
		// delete if all success
		go func() {
			defer graceful.Recover()
			if err := h.Service.RemoveMetadata(key); err != nil {
				hsLog.Errorf("delete-metadata %s fail: %s", key, err)
			}
		}()
	}
	hsLog.Infof("migration totally %d metadata and successed %d verions", total, sucNum)
	return
}

func (h *HashSlotService) migrateBuckets(stream pb.HashSlot_StreamingReceiveClient, edges hashslot.EdgeList) (errs []error) {
	var successBuckets int
	var migKeys [][]byte
	err := h.BucketService.Foreach(func(k []byte, _ []byte) error {
		if !hashslot.IsSlotInEdges(hashslot.CalcBytesSlot(k), edges) {
			return nil
		}
		migKeys = append(migKeys, k)
		return nil
	})
	total := len(migKeys)
	for len(migKeys) > 0 {
		k := migKeys[0]
		migKeys = migKeys[1:] // for earlier GC
		keyStr := util.BytesToStr(k)
		v, err := h.BucketService.GetBytes(keyStr)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		// send data
		if err := stream.Send(&pb.MigrationItem{
			Name: keyStr,
			Data: v,
			Dest: int32(entity.DestBucket),
		}); err != nil {
			hsLog.Debugf("send bucket %s err: %s", keyStr, err)
			errs = append(errs, err)
			continue
		}
		// recv response
		resp, err := stream.Recv()
		if err != nil {
			hsLog.Debugf("recv send-bucket %s response err: %s", keyStr, err)
			errs = append(errs, err)
			continue
		} else if !resp.Success {
			hsLog.Debugf("send-bucket %s recv failure resposne: %s", keyStr, resp.Message)
			errs = append(errs, errors.New(resp.Message))
			continue
		}
		successBuckets++
		go func() {
			defer graceful.Recover()
			if err := h.BucketService.Remove(keyStr); err != nil {
				hsLog.Errorf("delete-bucket %s err: %s", keyStr, err)
			}
		}()
	}
	if err != nil {
		errs = append(errs, err)
	}
	hsLog.Infof("migration totally %d buckets and successed %d", total, successBuckets)
	return
}
