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
	var total uint64
	for _, v := range capMap {
		total += v
	}
	avg := total / uint64(size)
	peerMap, err := logic.NewPeers().GetPeerMap()
	if err != nil {
		return nil, fmt.Errorf("get peers error: %w", err)
	}
	res := make(map[string]int64, len(peerMap))
	for k, v := range capMap {
		// skip self
		if k == pool.Config.Registry.ServerID {
			continue
		}
		if v = util.IfElse(join, v-avg, avg-v); v > 0 {
			info, ok := peerMap[k]
			if !ok {
				return nil, fmt.Errorf("unknown peers '%s'", k)
			}
			res[info.RpcAddress()] = int64(v)
		}
	}
	if len(res) == 0 {
		return nil, fmt.Errorf("non avaliable object servers")
	}
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
				stream := streamMap[cur]
				err = stream.Send(&pb.ObjectData{
					FileName: info.Name(),
					Data:     bt,
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
					// close stream before switch
					util.LogErr(stream.CloseSend())
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
	servs := pool.Discovery.GetServices(pool.Config.Discovery.MetaServName)
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
				if slices.StringsReplace(v.Locations, data.OriginLocate, newLoc) {
					versionsMap[ip] = append(versionsMap[ip], v)
				}
			}
		}(addr)
	}
	if err := dg.WaitUntilError(); err != nil {
		return err
	}
	// save file
	if err := Put(data.FileName, bytes.NewBuffer(data.Data)); err != nil {
		return err
	}
	// update locations
	//FIXME: partly update fails will cause inconsistent state
	failNum := atomic.NewInt32(0)
	var total int32
	for addr, versions := range versionsMap {
		dg.Todo()
		total += int32(len(versions))
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
	if fails := failNum.Load(); fails >= total/2 {
		return fmt.Errorf("too much failures when updating metadata (%d/%d)", fails, total)
	}
	return nil
}
