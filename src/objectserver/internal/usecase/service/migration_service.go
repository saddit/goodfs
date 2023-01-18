package service

import (
	"bytes"
	"common/cst"
	"common/datasize"
	"common/graceful"
	"common/logs"
	"common/pb"
	"common/util"
	"common/util/slices"
	"context"
	"fmt"
	"io"
	"io/fs"
	"math"
	"objectserver/internal/db"
	"objectserver/internal/usecase/logic"
	"objectserver/internal/usecase/pool"
	"objectserver/internal/usecase/webapi"
	"os"
	"path/filepath"
	"strings"

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
	clientMap := make(map[string]pb.ObjectMigrationClient, len(sizeMap))
	conns := make([]*grpc.ClientConn, 0, len(sizeMap))
	defer func() {
		for _, cc := range conns {
			util.LogErr(cc.Close())
		}
	}()
	ctx := context.Background()
	for k := range sizeMap {
		cc, err := grpc.Dial(k, grpc.WithInsecure())
		if err != nil {
			return err
		}
		conns = append(conns, cc)
		clientMap[k] = pb.NewObjectMigrationClient(cc)
	}
	next := make(chan string, 1)
	// open a goroutine to provide next server
	go func() {
		defer close(next)
		for k := range sizeMap {
			next <- k
		}
	}()
	successFlag := true
	cur := <-next
	leftSize := sizeMap[cur]
	curLocate := util.GetHostPort(pool.Config.Port)
	buf := make([]byte, datasize.MB*2)
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
		client, ok := clientMap[cur]
		if !ok {
			// switch to next
			logger.Errorf("not found client of %s", cur)
			cur = <-next
			return nil
		}
		stream, err := client.ReceiveData(ctx)
		if err != nil {
			// switch to next
			logger.Error("create stream to %s err: %s", cur, err)
			cur = <-next
			return nil
		}
		// open file
		func() {
			file, err := os.Open(path)
			if err != nil {
				// skip file
				successFlag = false
				logger.Errorf("open %s err: %s", path, err)
				return
			}
			defer file.Close()
			// send data
			func() {
				defer func() {
					resp, err := stream.CloseAndRecv()
					if err != nil {
						successFlag = false
						logger.Errorf("send file %s to %s fail, close and recv err: %s", info.Name(), cur, err)
					}
					if !resp.Success {
						successFlag = false
						logger.Errorf("send file %s to %s fail, close and recv message: %s", info.Name(), cur, resp.Message)
					}
				}()
				for {
					n, err := file.Read(buf)
					if err == io.EOF {
						break
					}
					if err != nil {
						logger.Errorf("read file %s err: %s", info.Name(), err)
						successFlag = false
						return
					}
					if err = stream.Send(&pb.ObjectData{
						FileName: info.Name(),
						Size:     info.Size(),
						Data:     buf[:n],
					}); err != nil {
						logger.Errorf("stream to %s interrupted, send %s data returns err: %s", cur, info.Name(), err)
						return
					}
				}
			}()

			
		}()

		// finish an object
		resp, err := client.ReceiveObject(ctx, &pb.ObjectInfo{
			FileName:     info.Name(),
			OriginLocate: curLocate,
		})
		if err != nil {
			logger.Errorf("finish send file %s err: %s", info.Name(), err)
			successFlag = false
			return nil
		}
		if !resp.Success {
			logger.Errorf("finish send file %s fail: %s", info.Name(), resp.Message)
			successFlag = false
			return nil
		}
		// switch to next server if already exceeds left size
		if leftSize -= info.Size(); leftSize <= 0 {
			cur = <-next
			leftSize = sizeMap[cur]
		}
		go func() {
			defer graceful.Recover()
			// remove local file
			if err = os.Remove(path); err != nil {
				logger.Errorf("migrate %s success, but delete fail: %s", path, err)
			}
		}()
		return nil
	})
	if err != nil {
		return err
	}
	return util.IfElse(successFlag, nil, fmt.Errorf("migration fail, see logs for detail"))
}

func (ms *MigrationService) Received(data *pb.ObjectInfo) error {
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
	// check file existed
	if !Exist(data.FileName) {
		return fmt.Errorf("file %s not exists", data.FileName)
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

func (ms *MigrationService) OpenFile(name string, size int64) (*os.File, error) {
	path := filepath.Join(pool.Config.StoragePath, name)
	stat, err := os.Stat(path)
	if err == nil {
		// if size equals, see as existed
		if stat.Size() == size {
			return nil, os.ErrExist
		}
		// some file may migrate failure. remove it if exist.
		if err = os.Remove(path); err != nil {
			return nil, err
		}
	}
	return os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, cst.OS.ModeUser)
}

func (ms *MigrationService) AppendData(file *os.File, data []byte) error {
	_, err := io.CopyBuffer(file, bytes.NewBuffer(data), make([]byte, 16*cst.OS.PageSize))
	return err
}
