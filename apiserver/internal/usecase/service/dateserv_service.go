package service

//TODO 更改从etcd获取数据节点
// 加入获取元数据节点的功能（主/从）

import (
	"apiserver/internal/entity"
	"apiserver/internal/usecase/pool"
	"apiserver/internal/usecase/selector"
	"common/util"
	"log"
	"time"
)

var dataServMap = util.NewSyncMap[string, entity.DataServ]()

func IsSuspendServer(srv *entity.DataServ) bool {
	return srv.GetState() == entity.ServStateSuspend ||
		srv.LastBeat.Add(pool.Config.Discovery.SuspendTimeout).Before(time.Now())
}

func IsDeadServer(srv *entity.DataServ) bool {
	return srv.GetState() == entity.ServStateDeath ||
		srv.LastBeat.Add(pool.Config.Discovery.DeadTimeout).Before(time.Now())
}

func IsAvailable(ip string) bool {
	var ds *entity.DataServ
	var ok bool
	if ds, ok = dataServMap.Get(ip); ok {
		return ds.IsAvailable()
	}
	return false
}

func ReceiveDataServer(ip string) {
	var ds *entity.DataServ
	var ok bool
	if ds, ok = dataServMap.Get(ip); ok {
		ds.SetState(entity.ServStateHealthy)
		ds.LastBeat = time.Now()
	} else {
		dataServMap.Put(ip, entity.NewDataServ(ip))
	}
}

func GetDataServers() []*entity.DataServ {
	//TODO 从注册中心获取
	ds := make([]*entity.DataServ, 0)
	CheckServerState()
	dataServMap.ForEach(func(key string, value *entity.DataServ) {
		ds = append(ds, value)
	})
	return ds
}

func SelectDataServer(balancer selector.Selector, size int) []string {
	ds := GetDataServers()
	if len(ds) == 0 {
		return []string{}
	}
	serv := make([]string, size)
	for i := 0; i < size; i++ {
		if len(ds) >= size-i {
			ds, serv[i] = balancer.Pop(ds)
		} else {
			serv[i] = balancer.Select(ds)
		}
	}
	return serv
}

func CheckServerState() {
	dataServMap.ForEach(func(key string, value *entity.DataServ) {
		if value == nil {
			dataServMap.Remove(key)
		} else if value.IsAvailable() {
			if IsSuspendServer(value) {
				value.SetState(entity.ServStateSuspend)
			}
		} else if IsDeadServer(value) {
			//第二次检查 未响应则移除
			log.Printf("Remove ip %v from dataServer map, last beat at %v\n", key, value.LastBeat)
			value.SetState(entity.ServStateDeath)
			go dataServMap.Remove(key)
		}
	})
}
