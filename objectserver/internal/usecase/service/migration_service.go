package service

import (
	"common/graceful"
	"common/logs"
	"common/pb"
	"common/util"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"io/fs"
	"objectserver/internal/db"
	"objectserver/internal/usecase/pool"
	"os"
	"path/filepath"
	"sync"
)

type MigrationService struct {
	CapacityDB *db.ObjectCapacity
}

func NewMigrationService(c *db.ObjectCapacity) *MigrationService {
	return &MigrationService{CapacityDB: c}
}

// DeviationValues calculate the required size of sending to or receiving from others depending on 'join'
func (ms *MigrationService) DeviationValues(join bool) (map[string]int64, error) {
	capMap, err := ms.CapacityDB.GetAll()
	if err != nil {
		return nil, err
	}
	size := util.IfElse(join, len(capMap)+1, len(capMap)-1)
	var total uint64
	for _, v := range capMap {
		total += v
	}
	avg := total / uint64(size)
	addrMap := pool.Discovery.GetServiceMapping(pool.Config.Registry.Name)
	res := make(map[string]int64, len(addrMap))
	for k, v := range capMap {
		if v = util.IfElse(join, v-avg, avg-v); v > 0 {
			addr := addrMap[k]
			res[addr] = int64(v)
		}
	}
	return res, nil
}

func (ms *MigrationService) SendingTo(sizeMap map[string]int64) error {
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
	errs := make(chan error)
	defer close(errs)
	// open a routine to log error
	go func() {
		logger := logs.New("migration-send")
		for err := range errs {
			logger.Error(err)
		}
	}()
	next := make(chan string, 1)
	// open a routine to provide next server
	go func() {
		defer close(next)
		for k := range sizeMap {
			next <- k
		}
	}()
	// record which server to send thread safe
	lock := sync.RWMutex{}
	cur := <-next
	leftSize := sizeMap[cur]
	errs <- filepath.Walk(pool.Config.StoragePath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
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
				errs <- err
				return
			}
			// send to server
			lock.RLock()
			stream := streamMap[cur]
			lock.RUnlock()
			err = stream.Send(&pb.ObjectData{
				FileName: info.Name(),
				Data:     bt,
			})
			if err != nil {
				errs <- fmt.Errorf("send %s to server fail: %w", path, err)
				return
			}
			// switch to next server if already exceeds left size
			lock.Lock()
			if leftSize -= info.Size(); leftSize <= 0 {
				cur = <-next
				leftSize = sizeMap[cur]
			}
			lock.Unlock()
			// remove local file
			if err = os.Remove(path); err != nil {
				errs <- fmt.Errorf("migrate %s success, but delete fail: %w", path, err)
			}
		}()
		return nil
	})
	wg.Wait()
	return nil
}

func (ms *MigrationService) Received(data *pb.ObjectData) error {
	return nil
}
