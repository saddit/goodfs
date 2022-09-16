package service

import (
	"common/util"
	"fmt"
	"metaserver/config"
	"metaserver/internal/usecase"
	"metaserver/internal/usecase/db"
	"metaserver/internal/usecase/logic"
)

//TODO Hash slot 服务执行
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
	info, exist, err := h.Store.Get(h.Cfg.ID)
	if err != nil || !exist {
		util.LogErrWithPre(
			"update slot info when leader change",
			util.IfElse[error](err == nil, usecase.ErrNotFound, err),
		)
		return
	}
	if isLeader {
		peers, err := logic.NewPeers().GetPeers()
		if err != nil {
			util.LogErr(err)
			return
		}
		var peersLoc []string
		for _, p := range peers {
			peersLoc = append(peersLoc, fmt.Sprint(p.Location, ":", p.HttpPort))
		}
		// if not enable raft, this will be empty
		if len(peersLoc) == 0 {
			peersLoc = append(peersLoc, h.HttpAddr)
		}
		info.Peers = peersLoc
		util.LogErr(h.Store.Save(h.Cfg.ID, info))
	}
}
