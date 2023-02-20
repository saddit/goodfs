package service

import (
	"bytes"
	"common/cst"
	"common/datasize"
	"common/graceful"
	"common/logs"
	"common/proto"
	"common/proto/pb"
	"common/response"
	"common/util"
	xmath "common/util/math"
	"common/util/slices"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math"
	"net/http"
	"objectserver/internal/db"
	"objectserver/internal/usecase/pool"
	"objectserver/internal/usecase/webapi"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"google.golang.org/grpc"
	"sync/atomic"
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
	servMap := pool.Discovery.GetServiceMapping(pool.Config.Registry.Name)
	res := make(map[string]int64, len(servMap))
	for k, v := range capMap {
		// skip self
		if k == pool.Config.Registry.ServerID {
			continue
		}
		if v = util.IfElse(join, v-avg, avg-v); v > 0 {
			rpcAddr, ok := servMap[k]
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
		var resp *pb.Response
		var inner error
		// if err is io.EOF, means stream has been closed by receiver
		if errors.Is(err, io.EOF) {
			err, resp = nil, new(pb.Response)
			inner = stream.RecvMsg(resp)
		} else {
			resp, inner = stream.CloseAndRecv()
		}
		if inner != nil {
			msLog.Errorf("send file %s fail, close and recv err: %s", name, inner)
			err = errors.Join(err, inner)
			return
		}
		if !resp.Success {
			err = errors.Join(err, fmt.Errorf("send file %s fail, close and recv message: %s", name, resp.Message))
		}
	}()
	var n int
	buf := make([]byte, xmath.MinNumber(size, int64(4*datasize.MB)))
	for {
		n, err = file.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read file %s err: %s", name, err)
		}
		if err = stream.Send(&pb.ObjectData{
			FileName: name,
			Size:     size,
			Data:     buf[:n],
		}); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			msLog.Errorf("stream interrupted, send %s data returns err: %s", name, err)
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
	if _, err = proto.ResolveResponse(client.FinishReceive(context.Background(), info)); err != nil {
		return fmt.Errorf("finish send file %s err: %w", info.FileName, err)
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
	// receive all errors
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
			// transfer file async
			dg.Todo()
			go func(toAddr string) {
				defer dg.Done()
				if inner := ms.sendFileTo(path, client, &pb.ObjectInfo{
					FileName:     info.Name(),
					Size:         info.Size(),
					OriginLocate: httpLocate,
				}); inner != nil {
					dg.Errors(inner)
					return
				}
				// remove file async if transfer success
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
			dg.Errors(err)
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
	newLoc, ok := pool.Discovery.GetService(pool.Config.Registry.Name, pool.Config.Registry.ServerID)
	if !ok {
		return fmt.Errorf("server unregister yet")
	}
	idx := strings.LastIndexByte(data.FileName, '.')
	seq := util.ToInt(data.FileName[idx+1:])
	hash := data.FileName[:idx]
	// get metadata locations of file
	var wg sync.WaitGroup
	failNum := atomic.Int32{}
	var total float64
	servs := pool.Discovery.GetServiceMapping(pool.Config.Discovery.MetaServName)
	if len(servs) == 0 {
		return fmt.Errorf("not exist meta-server")
	}
	for id, addr := range servs {
		if strings.HasSuffix(id, "slave") {
			continue
		}
		total++
		wg.Add(1)
		go func(ip string) {
			defer graceful.Recover()
			defer wg.Done()
			inner := webapi.UpdateVersionLocates(ip, hash, seq, newLoc)
			// ignore not found failure
			if response.CheckErrStatus(http.StatusNotFound, inner) {
				return
			}
			if inner != nil {
				logs.Std().Errorf("update locate of %s in %s fail: %s", data.FileName, ip, inner)
				failNum.Add(1)
			}
		}(addr)
	}
	wg.Wait()
	if fails := failNum.Load(); fails >= int32(math.Ceil(total/2)) {
		return fmt.Errorf("too much failures when updating metadata (%d/%.0f)", fails, total)
	}
	return nil
}

func (ms *MigrationService) OpenFile(name string, size int64) (*os.File, error) {
	path, ok := FindRealStoragePath(name)
	if !ok {
		path = filepath.Join(pool.DriverManager.SelectMountPointFallback(pool.Config.BaseMountPoint), pool.Config.StoragePath, name)
		util.LogErr(pool.PathDB.Put(name, path))
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
