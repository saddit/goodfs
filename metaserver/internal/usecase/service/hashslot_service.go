package service

import (
	"common/hashslot"
	"common/logs"
	"common/util"
	"context"
	"errors"
	"fmt"
	"github.com/tinylib/msgp/msgp"
	"google.golang.org/grpc"
	"metaserver/config"
	"metaserver/internal/entity"
	"metaserver/internal/usecase"
	"metaserver/internal/usecase/db"
	"metaserver/internal/usecase/pb"
	"metaserver/internal/usecase/pool"
	"time"
)

type HashSlotService struct {
	Store        *db.HashSlotDB
	Repo         usecase.IMetadataRepo
	Cfg          *config.HashSlotConfig
	startReceive func()
}

func NewHashSlotService(st *db.HashSlotDB, repo usecase.IMetadataRepo, cfg *config.HashSlotConfig) *HashSlotService {
	return &HashSlotService{
		Store:        st,
		Repo:         repo,
		Cfg:          cfg,
		startReceive: func() {},
	}
}

func (h *HashSlotService) OnLeaderChanged(isLeader bool) {
	if isLeader {
		var info *hashslot.SlotInfo
		var exist bool
		var err error
		if info, exist, err = h.Store.Get(h.Cfg.ID); !exist {
			if err != nil {
				util.LogErrWithPre("update slot info when leader changed", err)
				return
			}
			info = &hashslot.SlotInfo{Slots: h.Cfg.Slots}
		}
		info.Location = pool.HttpHostPort
		util.LogErr(h.Store.Save(h.Cfg.ID, info))
	}
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
	cc, err := grpc.Dial(fmt.Sprint(loc.GetHost(), ":", loc.GetRpcPort()))
	if err != nil {
		return err
	}
	client := pb.NewHashSlotClient(cc)
	resp, err := client.PrepareMigration(context.Background(), &pb.PrepareReq{
		Id: h.Cfg.ID,
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
	ctx, cancel := context.WithCancel(context.Background())
	h.startReceive = func() {
		cancel()
		h.startReceive = func() {}
	}
	go func() {
		select {
		case <-ctx.Done():
		case <-time.NewTicker(h.Cfg.PrepareTimeout).C:
			cancel()
			_ = h.Store.FinishMigrateFrom()
			logs.Std().Errorf("timeout migrating from %s", loc.GetHost())
		}
	}()
	return nil
}

func (h *HashSlotService) FinishReceiveItem(success bool) error {
	var newSlots []string
	if ok, _, slots := h.Store.GetMigrateFrom(); ok {
		newSlots = slots
	} else {
		return fmt.Errorf("get received slots fails: server is not in migrate-from")
	}
	if err := h.Store.FinishMigrateFrom(); err != nil {
		return err
	}
	if !success {
		return nil
	}
	info, _, err := h.Store.Get(h.Cfg.ID)
	if err != nil {
		return fmt.Errorf("update slot fails after finish migration: %w", err)
	}
	// combine to new slots, ignore error because both have been validated
	curEdges, _ := hashslot.WrapSlotsToEdges(info.Slots, info.Location)
	newEdges, _ := hashslot.WrapSlotsToEdges(newSlots, info.Location)
	info.Slots = hashslot.CombineEdges(curEdges, newEdges).Strings()
	// save new slot-info
	if err = h.Store.Save(h.Cfg.ID, info); err != nil {
		return fmt.Errorf("save new slot-info fails after finsih migrateion: %w", err)
	}
	return nil
}

func (h *HashSlotService) ReceiveItem(item *pb.MigrationItem) error {
	h.startReceive()
	logData := &entity.RaftData{
		Dest:     util.IfElse(item.IsVersion, entity.DestVersion, entity.DestMetadata),
		Type:     entity.LogInsert,
		Version:  &entity.Version{},
		Metadata: &entity.Metadata{},
	}
	if err := util.DecodeMsgp(
		util.IfElse[msgp.Unmarshaler](item.IsVersion, logData.Version, logData.Metadata),
		item.Data,
	); err != nil {
		return err
	}
	if ok, err := h.Repo.ApplyRaft(logData); ok && err != nil {
		return err
	} else if item.IsVersion {
		if err := h.Repo.AddVersion(item.GetName(), logData.Version); err != nil {
			return err
		}
	} else {
		if err := h.Repo.AddMetadata(logData.Metadata); err != nil {
			return err
		}
	}
	return nil
}

func (h *HashSlotService) AutoMigrate(toLoc *pb.LocationInfo, slots []string) error {
	if ok, host, _ := h.Store.GetMigrateTo(); !ok || host != toLoc.GetHost() {
		return fmt.Errorf("no ready to migrate to %s", toLoc.GetHost())
	}
	//TODO 如何确保过时的数据全部更新完毕: 由迁移双方自行控制，迁移成功后更新双方slot信息
	// 何时迁移: 指令触发时，指定 A to B with 10-100
	// 如何迁移：将K不在该服务器上的key通过RPC流服务迁移出去
	// 何时失败：迁移过程中发生异常中断
	// 合适成功：key-value全部迁移过去。etcd中的slots信息为迁移完成后的slot信息
	if err := h.Store.FinishMigrateTo(); err != nil {
		return err
	}
	// remove slots from current slot-info
	info, _, err := h.Store.Get(h.Cfg.ID)
	if err != nil {
		return fmt.Errorf("update slot fails after finish migration: %w", err)
	}
	curEdges, _ := hashslot.WrapSlotsToEdges(info.Slots, info.Location)
	delEdges, _ := hashslot.WrapSlotsToEdges(slots, info.Location)
	info.Slots = hashslot.RemoveEdges(curEdges, delEdges).Strings()
	// save new slot-info
	if err = h.Store.Save(h.Cfg.ID, info); err != nil {
		return fmt.Errorf("save new slot-info fails after finsih migrateion: %w", err)
	}
	return nil
}
