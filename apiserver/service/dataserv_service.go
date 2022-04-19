package service

import (
	"goodfs/apiserver/config"
	"goodfs/apiserver/model/dataserv"
	"goodfs/util"
	"log"
	"math/rand"
	"time"
)

var dataServMap = util.NewSyncMap[string, dataserv.DataServ]()

func IsSuspendServer(srv *dataserv.DataServ) bool {
	return srv.LastBeat.Add(config.SuspendTimeout * time.Second).Before(time.Now())
}

func IsDeadServer(srv *dataserv.DataServ) bool {
	return srv.LastBeat.Add(config.DeadTimeout * time.Second).Before(time.Now())
}

func IsAvailable(ip string) bool {
	var ds dataserv.DataServ
	if ok := dataServMap.Get2(ip, &ds); ok {
		return ds.IsAvailable()
	}
	return false
}

func ReceiveDataServer(ip string) {
	var ds dataserv.DataServ
	if ok := dataServMap.Get2(ip, &ds); ok {
		ds.State = dataserv.Healthy
		ds.LastBeat = time.Now()
	} else {
		dataServMap.Put(ip, dataserv.New(ip))
	}
}

func GetDataServers() []*dataserv.DataServ {
	ds := make([]*dataserv.DataServ, 0)
	dataServMap.ForEach(func(key string, value *dataserv.DataServ) {
		if value != nil {
			ds = append(ds, value)
		}
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
		if value.IsAvailable() {
			if IsSuspendServer(value) {
				value.State = dataserv.Suspend
			}
		} else if IsDeadServer(value) {
			//第二次检查 未响应则移除
			log.Printf("Remove ip %v from dataServer map\n", key)
			dataServMap.Remove(key)
		}
	})
}
