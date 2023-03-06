package logic

import (
	"adminserver/internal/entity"
	"adminserver/internal/usecase/db"
	"adminserver/internal/usecase/pool"
	"common/cst"
	"common/datasize"
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
	prefix := cst.EtcdPrefix.FmtSystemInfo(pool.Config.Discovery.Group, servName, "")
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
	for id, httpAddr := range pool.Discovery.GetServiceMapping(servName) {
		if s, ok := mp[id]; ok {
			s.HttpAddr = httpAddr
		}
	}
	return mp, nil
}

func (sm ServerMonitor) AliveCounts() map[string]int {
	counts := make(map[string]int, 3)
	counts[pool.Config.Discovery.ApiServName] = pool.Discovery.GetServiceCount(pool.Config.Discovery.ApiServName)
	counts[pool.Config.Discovery.MetaServName] = pool.Discovery.GetServiceCount(pool.Config.Discovery.MetaServName)
	counts[pool.Config.Discovery.DataServName] = pool.Discovery.GetServiceCount(pool.Config.Discovery.DataServName)
	return counts
}

func (sm ServerMonitor) NameOfServerNo(num int) string {
	var servName string
	switch num {
	case 0:
		servName = pool.Config.Discovery.ApiServName
	case 1:
		servName = pool.Config.Discovery.MetaServName
	case 2:
		servName = pool.Config.Discovery.DataServName
	}
	return servName
}

// StatTimeline cpu or mem stat timeline, statType = "cpu" | "mem", servNo = 0 | 1 | 2
func (sm ServerMonitor) StatTimeline(servNo int, statType string) map[string][]*db.TimeStat {
	tl := pool.StatDB.GetTimeline(sm.NameOfServerNo(servNo))
	res := make(map[string][]*db.TimeStat, len(tl))
	for k, v := range tl {
		res[k] = util.IfElse(statType == "cpu", v.CpuTimeline, v.MemTimeline)
	}
	return res
}

func (sm ServerMonitor) StatTimeLineOverview() (cpu map[string][]*db.TimeStat, mem map[string][]*db.TimeStat) {
	allNo := []int{0, 1, 2}
	cpu, mem = make(map[string][]*db.TimeStat, 0), make(map[string][]*db.TimeStat, 0)
	for _, v := range allNo {
		name := sm.NameOfServerNo(v)
		tl := pool.StatDB.GetTimeline(name)
		for _, v2 := range tl {
			cpu[name] = v2.CpuTimeline
			mem[name] = v2.MemTimeline
		}
	}
	return
}

func (ServerMonitor) EtcdStatus() ([]*entity.EtcdStatus, error) {
	ctx := context.Background()
	var arr []*entity.EtcdStatus
	for _, endpoint := range pool.Config.Etcd.Endpoint {
		resp, err := pool.Etcd.Status(ctx, endpoint)
		if err != nil {
			return nil, err
		}
		if resp.Errors == nil {
			resp.Errors = []string{}
		}
		arr = append(arr, &entity.EtcdStatus{
			DBSize:       datasize.DataSize(resp.DbSize),
			DBSizeInUse:  datasize.DataSize(resp.DbSizeInUse),
			AlarmMessage: resp.Errors,
			Endpoint:     endpoint,
			IsLearner:    resp.IsLearner,
		})
	}
	return arr, nil
}
