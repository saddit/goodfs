package logic

import (
	"adminserver/internal/entity"
	"adminserver/internal/usecase/db"
	"adminserver/internal/usecase/pool"
	"common/constrant"
	"common/system"
	"common/util"
	"context"
	"strings"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type ServerMonitor struct{}

func NewServerMonitor() *ServerMonitor {
	return new(ServerMonitor)
}

func (ServerMonitor) SysInfo(servName string) (map[string]*system.Info, error) {
	prefix := constrant.EtcdPrefix.FmtSystemInfo(pool.Config.Discovery.Group, servName, "")
	resp, err := pool.Etcd.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	mp := make(map[string]*system.Info, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var info system.Info
		if err = util.DecodeMsgp(&info, kv.Value); err != nil {
			return nil, err
		}
		sp := strings.Split(string(kv.Key), "/")
		serverId := sp[len(sp)-1]
		mp[serverId] = &info
	}
	return mp, nil
}

func (sm ServerMonitor) ServerStat(servName string) (map[string]*entity.ServerInfo, error) {
	mp := make(map[string]*entity.ServerInfo)
	sysMap, err := sm.SysInfo(servName)
	if err != nil {
		return nil, err
	}
	for id, sysInfo := range sysMap {
		mp[id] = &entity.ServerInfo{SysInfo: sysInfo, ServerID: id}
	}
	for id, httpAddr := range pool.Discovery.GetServiceMapping(servName, false) {
		mp[id].HttpAddr = httpAddr
	}
	for id, rpcAddr := range pool.Discovery.GetServiceMapping(servName, true) {
		mp[id].RpcAddr = rpcAddr
	}
	return mp, nil
}

// StatTimeline cpu or mem stat timeline, statType = "cpu" | "mem", servNo = 0 | 1 | 2
func (sm ServerMonitor) StatTimeline(servNo int, statType string) map[string][]*db.TimeStat {
	var servName string
	switch servNo {
	case 0:
		servName = pool.Config.Discovery.ApiServName
	case 1:
		servName = pool.Config.Discovery.MetaServName
	case 2:
		servName = pool.Config.Discovery.DataServName
	}
	tl := pool.StatDB.GetTimeline(servName)
	res := make(map[string][]*db.TimeStat, len(tl))
	for k, v := range tl {
		res[k] = util.IfElse(statType == "cpu", v.CpuTimeline, v.MemTimeline)
	}
	return res
}
