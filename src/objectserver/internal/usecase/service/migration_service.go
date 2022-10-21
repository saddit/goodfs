package service

import (
	"bytes"
	"common/graceful"
	"common/logs"
	"common/pb"
	"common/util"
	"common/util/slices"
	"context"
	"fmt"
	"io/fs"
	"math"
	"objectserver/internal/db"
	"objectserver/internal/usecase/logic"
	"objectserver/internal/usecase/pool"
	"objectserver/internal/usecase/webapi"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"go.uber.org/atomic"
	"google.golang.org/grpc"
)

type MigrationService struct {
	CapacityDB *db.ObjectCapacity
}

func NewMigrationService(c *db.ObjectCapacity) *MigrationService {
	return &MigrationService{CapacityDB: c}
}

// DeviationValues calculate the required size of sending to or receiving from others depending on 'join'
// return map(key=rpc-addr,value=capacity)
func (ms *MigrationService) DeviationValues(join bool) (map[string]int64, error) {
	capMap, err := ms.CapacityDB.GetAll()
	if err != nil {
		return nil, err
	}
	size := util.IfElse(join, len(capMap)+1, len(capMap)-1)
	if size == 0 {
		return nil, fmt.Errorf("non avaliable object servers")
	}
	var total float64
	for _, v := range capMap {
		total += float64(v)
	}
	avg := uint64(math.Ceil(total / float64(size)))
	rpcMap := logic.NewPeers().GetPeerMap()
	res := make(map[string]int64, len(rpcMap))
	for k, v := range capMap {
		// skip self
		if k == pool.Config.Registry.ServerID {
			continue
		}
		if v = util.IfElse(join, v-avg, avg-v); v > 0 {
			rpcAddr, ok := rpcMap[k]
			if !ok {
				return nil, fmt.Errorf("unknown peers '%s'", k)
			}
			res[rpcAddr] = int64(v)
		}
	}
	if len(res) == 0 {
		return nil, fmt.Errorf("non avaliable object servers")
	}
	logs.Std().Debugf("DeviationValues: %+v", res)
	return res, nil
}

func (ms *MigrationService) SendingTo(sizeMap map[string]int64) error {
	logger := logs.New("migration-send")
	// dail connection to all servers
	streamMap := make(map[string]pb.ObjectMigration_ReceiveObjectClient, len(sizeMap))
	conns := make([]*grpc.ClientConn, 0, len(sizeMap))
	defer func() {
		for _, cc := range conns {
			util.LogErr(cc.Close())
		}
	}()
	for k := range sizeMap {
		cc, err := grpc.Dial(k, grpc.WithInsecure())
		if err != nil {
			return err
		}
		conns = append(conns, cc)
		stream, err := pb.NewObjectMigrationClient(cc).ReceiveObject(context.Background())
		if err != nil {
			return err
		}
		streamMap[k] = stream
	}
	defer func() {
		for _, stream := range streamMap {
			util.LogErr(stream.CloseSend())
		}
	}()
	// sending data concurrency, limiting num of routine under 16
	wg := sync.WaitGroup{}
	ctrl := make(chan struct{}, 16)
	defer close(ctrl)
	next := make(chan string, 1)
	// open a routine to provide next server
	go func() {
		defer close(next)
		for k := range sizeMap {
			next <- k
		}
	}()
	// record which server to send thread safe
	successFlag := true
	lock := sync.Mutex{}
	cur := <-next
	leftSize := sizeMap[cur]
	curLocate := util.GetHostPort(pool.Config.Port)
	//TODO(perf): run multi goroutines to each server
	// race files by CAS on memory
	err := filepath.Walk(pool.Config.StoragePath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if cur == "" {
			return fmt.Errorf("server is depleted but there are still files not being transferred")
		}
		ctrl <- struct{}{}
		wg.Add(1)
		go func() {
			defer graceful.Recover()
			defer func() {
				wg.Done()
				<-ctrl
			}()
			// read data
			bt, err := os.ReadFile(path)
			if err != nil {
				logger.Errorf("read file %s err: %s", path, err)
				return
			}
			// send to server
			if ok := func() bool {
				lock.Lock()
				defer lock.Unlock()
				stream, ok := streamMap[cur]
				if !ok {
					logger.Errorf("not found stream of %s", cur)
					return false
				}
				err = stream.Send(&pb.ObjectData{
					FileName:     info.Name(),
					Data:         bt,
					OriginLocate: curLocate,
				})
				if err != nil {
					logger.Errorf("send %s to server %s fail: %s", path, cur, err)
					return false
				}
				resp, err := stream.Recv()
				if err != nil {
					logger.Errorf("send %s to %s server fail: %s", path, cur, err)
					return false
				}
				if !resp.Success {
					logger.Errorf("send %s to server %s fail: %s", path, cur, resp.Message)
					return false
				}
				// switch to next server if already exceeds left size
				if leftSize -= info.Size(); leftSize <= 0 {
					cur = <-next
					leftSize = sizeMap[cur]
				}
				return true
			}(); !ok {
				successFlag = false
				return
			}
			// remove local file
			if err = os.Remove(path); err != nil {
				successFlag = false
				logger.Errorf("migrate %s success, but delete fail: %s", path, err)
				return
			}
		}()
		return nil
	})
	wg.Wait()
	if err != nil {
		return err
	}
	return util.IfElse(successFlag, nil, fmt.Errorf("migration fail, see logs for detail"))
}

func (ms *MigrationService) Received(data *pb.ObjectData) error {
	servs := pool.Discovery.GetServices(pool.Config.Discovery.MetaServName, false)
	if len(servs) == 0 {
		return fmt.Errorf("not exist meta-server")
	}
	newLoc := util.GetHostPort(pool.Config.Port)
	hash := strings.Split(data.FileName, ".")[0]
	versionsMap := make(map[string][]*pb.Version)
	// get metadata locations of file
	dg := util.NewDoneGroup()
	defer dg.Close()
	for _, addr := range servs {
		dg.Todo()
		go func(ip string) {
			defer graceful.Recover()
			defer dg.Done()
			versions, err := webapi.VersionsByHash(ip, hash)
			if err != nil {
				dg.Error(err)
				return
			}
			for _, v := range versions {
				// index sensitive: locations[i] to shard[i]
				sp := strings.Split(data.FileName, ".")
				seq := util.ToInt(slices.Last(sp))
				// do not change if location of this shard has been updated
				if v.Locations[seq] == data.OriginLocate {
					v.Locations[seq] = newLoc
					versionsMap[ip] = append(versionsMap[ip], v)
				}
			}
		}(addr)
	}
	if err := dg.WaitUntilError(); err != nil {
		return err
	}
	// if no metadata exist for this file from 'origin-locate', deprecate it.
	if len(versionsMap) == 0 {
		logs.Std().Infof("deprecated: non metadata needs to update for %s (from %s)", data.FileName, data.OriginLocate)
		return nil
	}
	// save file
	if err := Put(data.FileName, bytes.NewBuffer(data.Data)); err != nil {
		return err
	}
	// update locations
	//FIXME: this will cause inconsistent state
	failNum := atomic.NewInt32(0)
	var total float64
	for addr, versions := range versionsMap {
		dg.Todo()
		total += float64(len(versions))
		go func(ip string, arr []*pb.Version) {
			defer graceful.Recover()
			defer dg.Done()
			for _, ver := range arr {
				if err := webapi.UpdateVersionLocates(ip, ver.Name, int(ver.Sequence), ver.Locations); err != nil {
					logs.Std().Errorf("update %+v to meta-server %s fail: %s", ver, ip, err)
					failNum.Inc()
				}
			}
		}(addr, versions)
	}
	dg.Wait()
	if fails := failNum.Load(); fails >= int32(math.Ceil(total/2.0)) {
		return fmt.Errorf("too much failures when updating metadata (%d/%.0f)", fails, total)
	}
	return nil
}
