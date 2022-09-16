package service

import (
	"common/hashslot"
	"common/util"
	"metaserver/config"
	"metaserver/internal/usecase/db"
)

type HashSlotService struct {
	Store    *db.HashSlotDB
	Cfg      *config.HashSlotConfig
	HttpAddr string
}

func NewHashSlotService(st *db.HashSlotDB, cfg *config.HashSlotConfig, httpAddr string) *HashSlotService {
	return &HashSlotService{
		Store:    st,
		Cfg:      cfg,
		HttpAddr: httpAddr,
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
		info.Location = h.HttpAddr
		util.LogErr(h.Store.Save(h.Cfg.ID, info))
	}
}

// AutoMigrate 迁移数据
// TODO 如何确保过时的数据全部更新完毕: 由迁移双方自行控制，迁移成功后更新双方slot信息
// TODO 何时迁移: 指令触发时，指定 A to B with 10-100
// TODO 如何迁移：将K不在该服务器上的key通过RPC流服务迁移出去，也就是说我需要编写HashSlotRpcServer
// TODO 何时失败：10-100不完全属于A、A或B繁忙、迁移过程中发生异常中断
// TODO 合适成功：key-value全部迁移过去。etcd中的slots信息为迁移完成后的slot信息
func (h *HashSlotService) AutoMigrate(toLoc string, slots []string) error {
	return nil
}
