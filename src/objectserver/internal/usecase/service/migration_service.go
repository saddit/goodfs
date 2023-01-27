package service

import (
	"bytes"
	"common/cst"
	"common/datasize"
	"common/graceful"
	"common/logs"
	"common/pb"
	"common/util"
	xmath "common/util/math"
	"common/util/slices"
	"context"
	"fmt"
	"io"
	"io/fs"
	"math"
	"objectserver/internal/db"
	"objectserver/internal/usecase/pool"
	"objectserver/internal/usecase/webapi"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/atomic"
	"google.golang.org/grpc"
)

var (
	msLog = logs.New("migration-service")
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
	size := util.IfElse(join, len(capMap), len(capMap)-1)
	if size == 0 {
		return nil, fmt.Errorf("non avaliable object servers")
	}
	var total float64
	for _, v := range capMap {
		total += float64(v)
	}
	avg := uint64(math.Ceil(total / float64(size)))
	rpcMap := pool.Discovery.GetServiceMapping(pool.Config.Registry.Name, true)
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

func (ms *MigrationService) writeStream(stream pb.ObjectMigration_ReceiveDataClient, file io.Reader, name string, size int64) (err error) {
	defer func() {
		resp, inner := stream.CloseAndRecv()
		if inner != nil {
			msLog.Errorf("send file %s fail, close and recv err: %s", name, inner)
		}
		if !resp.Success {
			err = fmt.Errorf("send file %s fail, close and recv message: %s", name, resp.Message)
		}
	}()
	buf := make([]byte, xmath.MinNumber(size, int64(4*datasize.MB)))
	for {
		n, inner := file.Read(buf)
		if inner == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read file %s err: %s", name, inner)
		}
		if inner = stream.Send(&pb.ObjectData{
			FileName: name,
			Size:     size,
			Data:     buf[:n],
		}); inner != nil {
			msLog.Errorf("stream interrupted, send %s data returns err: %s", name, inner)
			return
		}
	}
	return
}

func (ms *MigrationService) sendFileTo(path string, client pb.ObjectMigrationClient, info *pb.ObjectInfo) error {
	// open stream
	stream, err := client.ReceiveData(context.Background())
	if err != nil {
		return fmt.Errorf("create stream err: %w", err)
	}
	// open file
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open %s err: %w", path, err)
	}
	defer file.Close()
	// send data
	if err = ms.writeStream(stream, file, info.FileName, info.Size); err != nil {
		return err
	}
	// finish an object
	resp, err := client.FinishReceive(context.Background(), info)
	if err != nil {
		return fmt.Errorf("finish send file %s err: %w", info.FileName, err)
	}
	if !resp.Success {
		return fmt.Errorf("finish send file %s fail: %s", info.FileName, resp.Message)
	}
	return nil
}

func (ms *MigrationService) SendingTo(httpLocate string, sizeMap map[string]int64) error {
	if len(sizeMap) == 0 {
		return nil
	}
	clientMap := make(map[string]pb.ObjectMigrationClient, len(sizeMap))
	conns := make([]*grpc.ClientConn, 0, len(sizeMap))
	addrs := make([]string, 0, len(sizeMap))
	defer func() {
		for _, cc := range conns {
			util.LogErr(cc.Close())
		}
	}()
	for k := range sizeMap {
		// dail connection to all servers
		cc, err := grpc.Dial(k, grpc.WithInsecure())
		if err != nil {
			return err
		}
		conns = append(conns, cc)
		addrs = append(addrs, k)
		clientMap[k] = pb.NewObjectMigrationClient(cc)
	}
	success := true
	dg := util.LimitDoneGroup(128)
	defer dg.Close()
	go func() {
		defer graceful.Recover()
		for err := range dg.WaitError() {
			msLog.Error(err)
			success = false
		}
	}()
	cur := slices.First(addrs)
	slices.RemoveFirst(&addrs)
	leftSize := sizeMap[cur]
	var err error
	for _, mp := range pool.DriverManager.GetAllMountPoint() {
		err = filepath.Walk(filepath.Join(mp, pool.Config.StoragePath), func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				msLog.Errorf("walk path %s err: %s, will skip this path", path, err)
				return nil
			}
			if info.IsDir() {
				return nil
			}
			// switch to next server if already exceeds left size
			if leftSize -= info.Size(); leftSize <= 0 {
				if len(addrs) > 0 {
					cur = slices.First(addrs)
					slices.RemoveFirst(&addrs)
					leftSize = sizeMap[cur]
				} else {
					return io.EOF
				}
			}
			client := clientMap[cur]
			dg.Todo()
			go func(toAddr string) {
				defer dg.Done()
				if inner := ms.sendFileTo(path, client, &pb.ObjectInfo{
					FileName:     info.Name(),
					Size:         info.Size(),
					OriginLocate: httpLocate,
				}); inner != nil {
					dg.Error(inner)
					return
				}
				go func() {
					defer graceful.Recover()
					// remove local file
					if inner := os.Remove(path); inner != nil {
						msLog.Errorf("migrate %s success, but delete fail: %s", path, err)
					}
				}()
			}(cur)
			return nil
		})
		if err == io.EOF {
			break
		}
		if err != nil {
			dg.Error(err)
		}
	}
	dg.Wait()
	if err != nil && err != io.EOF {
		return err
	}
	return util.IfElse(success, nil, fmt.Errorf("migration fail, see logs for detail"))
}

func (ms *MigrationService) FinishObject(data *pb.ObjectInfo) error {
	// check file existed
	realPath, ok := FindRealStoragePath(data.FileName)
	if !ok {
		return fmt.Errorf("file %s not exists", data.FileName)
	}
	statInfo, err := os.Stat(realPath)
	if err != nil {
		return fmt.Errorf("file %s not exists: %w", data.FileName, err)
	}
	if statInfo.Size() != data.Size {
		return fmt.Errorf("file %s size excpect %d, actual %d", data.FileName, data.Size, statInfo.Size())
	}
	// add to capacity db
	pool.ObjectCap.AddCap(data.Size)
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
				idx := strings.LastIndexByte(data.FileName, '.')
				seq := util.ToInt(data.FileName[idx+1:])
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
	if fails := failNum.Load(); fails >= int32(math.Ceil(total/2)) {
		return fmt.Errorf("too much failures when updating metadata (%d/%.0f)", fails, total)
	}
	return nil
}

func (ms *MigrationService) OpenFile(name string, size int64) (*os.File, error) {
	path, ok := FindRealStoragePath(name)
	if !ok {
		path = filepath.Join(pool.DriverManager.SelectMountPointFallback(pool.Config.BaseMountPoint), pool.Config.StoragePath, name)
		_ = pool.PathDB.Put(name, path)
	}
	stat, err := os.Stat(path)
	if err == nil {
		// if size equals, see as existed
		if stat.Size() == size {
			return nil, os.ErrExist
		}
		// some file may migrate failure. remove it if exists.
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
