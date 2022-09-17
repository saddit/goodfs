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
	return h.Store.ReadyMigrateTo(loc.GetHost(), slots)
}

func (h *HashSlotService) PrepareMigrationFrom(loc *pb.LocationInfo, slots []string) error {
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
		case <-time.NewTicker(10 * time.Second).C:
			cancel()
			_ = h.Store.FinishMigrateFrom(false)
			logs.Std().Errorf("timeout migrating from %s", loc.GetHost())
		}
	}()
	return nil
}

func (h *HashSlotService) FinishReceiveItem(ok bool) error {
	return h.Store.FinishMigrateFrom(ok)
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
	if ok, host := h.Store.GetMigrateTo(); !ok || host != toLoc.GetHost() {
		return fmt.Errorf("no ready to migrate to %s", toLoc.GetHost())
	}
	// TODO 如何确保过时的数据全部更新完毕: 由迁移双方自行控制，迁移成功后更新双方slot信息
	// TODO 何时迁移: 指令触发时，指定 A to B with 10-100
	// TODO 如何迁移：将K不在该服务器上的key通过RPC流服务迁移出去，也就是说我需要编写HashSlotRpcServer
	// TODO 何时失败：10-100不完全属于A、A或B繁忙、迁移过程中发生异常中断
	// TODO 合适成功：key-value全部迁移过去。etcd中的slots信息为迁移完成后的slot信息
	return h.Store.FinishMigrateTo(true)
}
