package db

import (
	"bytes"
	"common/constrant"
	"common/graceful"
	"common/logs"
	"common/system"
	"common/util"
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

var (
	statLog     = logs.New("stat-db")
	maxTimeStat = 60
)

type TimeStat struct {
	Time    time.Time
	Percent float64
}

type statTimeline struct {
	CpuTimeline []*TimeStat
	MemTimeline []*TimeStat
}

func newStatTimeline() statTimeline {
	return statTimeline{
		CpuTimeline: make([]*TimeStat, 0, 60),
		MemTimeline: make([]*TimeStat, 0, 60),
	}
}

func (st *statTimeline) Append(cpu *TimeStat, mem *TimeStat) {
	// limit size
	if len(st.CpuTimeline) == maxTimeStat {
		st.CpuTimeline = st.CpuTimeline[1:maxTimeStat]
	}
	if len(st.MemTimeline) == maxTimeStat {
		st.MemTimeline = st.CpuTimeline[1:maxTimeStat]
	}
	// add new stat
	st.CpuTimeline = append(st.CpuTimeline, cpu)
	st.MemTimeline = append(st.MemTimeline, mem)
}

type ServerStatDB struct {
	Cli       clientv3.Watcher
	GroupName string
	Services  []string
	closeFn   func()
	timeline  map[string]map[string]statTimeline
}

func NewServerStatDB(cli clientv3.Watcher, groupName string, services []string) *ServerStatDB {
	o := &ServerStatDB{
		Cli:       cli,
		GroupName: groupName,
		Services:  services,
		timeline:  map[string]map[string]statTimeline{},
	}
	o.init()
	return o
}

func (sdb *ServerStatDB) GetTimeline(servName string) map[string]statTimeline {
	if tls, ok := sdb.timeline[servName]; ok {
		return tls
	}
	return map[string]statTimeline{}
}

func (sdb *ServerStatDB) init() {
	ctx, cancel := context.WithCancel(context.Background())
	sdb.closeFn = cancel
	for _, v := range sdb.Services {
		ch := sdb.Cli.Watch(ctx, constrant.EtcdPrefix.FmtSystemInfo(sdb.GroupName, v, ""))
		go sdb.watching(v, ch)
	}
	return
}

func (sdb *ServerStatDB) watching(serv string, ch clientv3.WatchChan) {
	defer graceful.Recover()
	sdb.timeline[serv] = map[string]statTimeline{}
	mp := sdb.timeline[serv]
	for v := range ch {
		ts := time.Now()
		for _, event := range v.Events {
			idx := bytes.LastIndex(event.Kv.Key, constrant.EtcdPrefix.Sep)
			id := string(event.Kv.Key[idx+1:])
			var sysInfo system.Info
			if err := util.DecodeMsgp(&sysInfo, event.Kv.Value); err != nil {
				statLog.Error(err)
				continue
			}
			tl, ok := mp[id]
			if !ok {
				mp[id] = newStatTimeline()
				tl = mp[id]
			}
			tl.Append(&TimeStat{
				Time:    ts,
				Percent: sysInfo.CpuStatus.UsedPercent,
			}, &TimeStat{
				Time:    ts,
				Percent: float64(sysInfo.MemStatus.Used) / float64(sysInfo.MemStatus.All),
			})
		}
	}
}

func (sdb *ServerStatDB) Close() error {
	sdb.closeFn()
	return nil
}
