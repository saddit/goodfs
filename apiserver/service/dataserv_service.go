package service

import (
	"goodfs/apiserver/global"
	"goodfs/apiserver/model/dataserv"
	"goodfs/lib/util"
	"log"
	"math/rand"
	"time"
)

var dataServMap = util.NewSyncMap[string, dataserv.DataServ]()

func IsSuspendServer(srv *dataserv.DataServ) bool {
	return srv.GetState() == dataserv.Suspend ||
		srv.LastBeat.Add(global.Config.SuspendTimeout).Before(time.Now())
}

func IsDeadServer(srv *dataserv.DataServ) bool {
	return srv.GetState() == dataserv.Death ||
		srv.LastBeat.Add(global.Config.DeadTimeout).Before(time.Now())
}

func IsAvailable(ip string) bool {
	var ds *dataserv.DataServ
	var ok bool
	if ds, ok = dataServMap.Get(ip); ok {
		return ds.IsAvailable()
	}
	return false
}

func ReceiveDataServer(ip string) {
	var ds *dataserv.DataServ
	var ok bool
	if ds, ok = dataServMap.Get(ip); ok {
		ds.SetState(dataserv.Healthy)
		ds.LastBeat = time.Now()
	} else {
		dataServMap.Put(ip, dataserv.New(ip))
	}
}

func GetDataServers() []*dataserv.DataServ {
	ds := make([]*dataserv.DataServ, 0)
	CheckServerState()
	dataServMap.ForEach(func(key string, value *dataserv.DataServ) {
		ds = append(ds, value)
	})
	return ds
}

func RandomDataServer() (string, bool) {
	ds := GetDataServers()
	size := len(ds)
	if size == 0 {
		return "", false
	}
	return ds[rand.Intn(size)].Ip, true
}

func CheckServerState() {
	dataServMap.ForEach(func(key string, value *dataserv.DataServ) {
		if value == nil {
			dataServMap.Remove(key)
		} else if value.IsAvailable() {
			if IsSuspendServer(value) {
				value.SetState(dataserv.Suspend)
			}
		} else if IsDeadServer(value) {
			//第二次检查 未响应则移除
			log.Printf("Remove ip %v from dataServer map, last beat at %v\n", key, value.LastBeat)
			value.SetState(dataserv.Death)
			go dataServMap.Remove(key)
		}
	})
}
