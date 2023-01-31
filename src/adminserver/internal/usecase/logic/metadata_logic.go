package logic

import (
	"adminserver/internal/entity"
	"adminserver/internal/usecase/pool"
	"adminserver/internal/usecase/webapi"
	"common/collection/set"
	"common/cst"
	"common/hashslot"
	"common/logs"
	"common/pb"
	"common/response"
	"common/util"
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"sync"

	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type MetadataCond struct {
	Name     string `form:"name"`
	Bucket   string `form:"bucket"`
	Version  int    `form:"version"`
	Page     int    `form:"page" binding:"required"`
	PageSize int    `form:"pageSize" binding:"required"`
	OrderBy  string `form:"orderBy"`
	Desc     bool   `form:"desc"`
}

type BucketCond struct {
	Name     string `form:"name"`
	Page     int    `form:"page" binding:"required"`
	PageSize int    `form:"pageSize" binding:"required"`
	OrderBy  string `form:"orderBy"`
	Desc     bool   `form:"desc"`
}

type Metadata struct{}

func NewMetadata() *Metadata {
	return new(Metadata)
}

func (m *Metadata) MetadataPaging(cond MetadataCond) ([]*entity.Metadata, int, error) {
	servers := pool.Discovery.GetServices(pool.Config.Discovery.MetaServName, false)
	lst := make([]*entity.Metadata, 0, len(servers)*cond.Page*cond.PageSize)
	if len(servers) == 0 {
		logs.Std().Warn("not found any metadata server")
		return lst, 0, nil
	}
	mux := sync.Mutex{}
	var totals int
	dg := util.NewDoneGroup()
	defer dg.Close()
	for _, ip := range servers {
		dg.Add(1)
		go func(loc string) {
			defer dg.Done()
			key := util.IfElse(cond.Bucket == "", cond.Name, fmt.Sprint(cond.Bucket, "/", cond.Name))
			data, total, err := webapi.ListMetadata(loc, key, cond.Page*cond.PageSize)
			if err != nil {
				dg.Error(err)
				return
			}
			mux.Lock()
			defer mux.Unlock()
			totals += total
			lst = append(lst, data...)
		}(ip)
	}
	if err := dg.WaitUntilError(); err != nil {
		return nil, 0, err
	}
	// TODO: remove order logic
	if st, ed, ok := util.PagingOffset(cond.Page, cond.PageSize, len(lst)); ok {
		sort.Slice(lst, func(i, j int) bool {
			var res bool
			switch cond.OrderBy {
			default:
				fallthrough
			case "create_time":
				res = lst[i].CreateTime < lst[j].CreateTime
			case "update_time":
				res = lst[i].UpdateTime < lst[j].UpdateTime
			case "name":
				res = lst[i].Name < lst[j].Name
			}
			return util.IfElse(cond.Desc, !res, res)
		})
		return lst[st:ed], totals, nil
	}
	return []*entity.Metadata{}, 0, nil
}

func (m *Metadata) VersionPaging(cond MetadataCond, token string) ([]byte, int, error) {
	return webapi.ListVersion(SelectApiServer(), cond.Name, cond.Bucket, cond.Page, cond.PageSize, token)
}

func (m *Metadata) BucketPaging(cond *BucketCond) ([]*entity.Bucket, int, error) {
	servers := pool.Discovery.GetServices(pool.Config.Discovery.MetaServName, false)
	lst := make([]*entity.Bucket, 0, len(servers)*cond.Page*cond.PageSize)
	if len(servers) == 0 {
		logs.Std().Warn("not found any metadata server")
		return lst, 0, nil
	}
	var totals int
	mux := &sync.Mutex{}
	dg := util.NewDoneGroup()
	defer dg.Close()
	for _, ip := range servers {
		dg.Add(1)
		go func(loc string) {
			defer dg.Done()
			data, total, err := webapi.ListBuckets(loc, cond.Name, cond.Page*cond.PageSize)
			if err != nil {
				dg.Error(err)
				return
			}
			mux.Lock()
			defer mux.Unlock()
			totals += total
			lst = append(lst, data...)
		}(ip)
	}
	if err := dg.WaitUntilError(); err != nil {
		return nil, 0, err
	}
	st, ed, _ := util.PagingOffset(cond.Page, cond.PageSize, len(lst))
	return lst[st:ed], totals, nil
}

func (m *Metadata) StartMigration(srcID, destID string, slots []string) error {
	mp := pool.Discovery.GetServiceMapping(pool.Config.Discovery.MetaServName, true)
	srcAddr, destAddr := mp[srcID], mp[destID]
	if srcAddr == "" || destAddr == "" {
		return response.NewError(400, "invalid server id")
	}
	cc, err := grpc.Dial(srcAddr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	cli := pb.NewHashSlotClient(cc)
	host, port, err := net.SplitHostPort(destAddr)
	if err != nil {
		return err
	}
	resp, err := cli.StartMigration(context.Background(), &pb.MigrationReq{
		Slots: slots,
		TargetLocation: &pb.LocationInfo{
			Host:    host,
			RpcPort: port,
		},
	})
	if err != nil {
		return err
	}
	if !resp.Success {
		return response.NewError(400, resp.GetMessage())
	}
	return nil
}

func (m *Metadata) GetSlotsDetail() (map[string]*hashslot.SlotInfo, error) {
	prefix := cst.EtcdPrefix.FmtHashSlot(pool.Config.Discovery.Group, pool.Config.Discovery.MetaServName, "")
	resp, err := pool.Etcd.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	res := make(map[string]*hashslot.SlotInfo, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var info hashslot.SlotInfo
		if err = util.DecodeMsgp(&info, kv.Value); err != nil {
			return nil, err
		}
		// to avoid NULL in json
		if info.Slots == nil {
			info.Slots = []string{}
		}
		res[info.GroupID] = &info
	}
	return res, nil
}

func (m *Metadata) GetMasterServerIds() set.Set {
	mp := pool.Discovery.GetServiceMappingWith(pool.Config.Discovery.MetaServName, false, true)
	masters := set.OfMapKeys(mp)
	return masters
}

func (m *Metadata) LeaveRaftCluster(servId string) error {
	mp := pool.Discovery.GetServiceMapping(pool.Config.Discovery.MetaServName, true)
	servAddr, ok := mp[servId]
	if !ok {
		return response.NewError(400, "unknown server id")
	}
	cc, err := grpc.Dial(servAddr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	_, err = pb.NewRaftCmdClient(cc).LeaveCluster(context.Background(), new(pb.EmptyReq))
	return err
}

func (m *Metadata) JoinRaftCluster(masterId, servId string) error {
	mp := pool.Discovery.GetServiceMapping(pool.Config.Discovery.MetaServName, true)
	servAddr, ok := mp[servId]
	if !ok {
		return response.NewError(400, "unknown server id")
	}
	cc, err := grpc.Dial(servAddr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	client := pb.NewRaftCmdClient(cc)
	resp, err := client.JoinLeader(context.Background(), &pb.JoinLeaderReq{MasterId: masterId})
	if err != nil {
		return err
	}
	if !resp.Success {
		return response.NewError(400, resp.Message)
	}
	return nil
}

func (m *Metadata) GetPeers(servId string) ([]*entity.ServerInfo, error) {
	mp := pool.Discovery.GetServiceMapping(pool.Config.Discovery.MetaServName, true)
	servAddr, ok := mp[servId]
	if !ok {
		return nil, response.NewError(400, "unknown server id")
	}
	cc, err := grpc.Dial(servAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	resp, err := pb.NewMetadataApiClient(cc).GetPeers(context.Background(), new(pb.EmptyReq))
	if err != nil {
		return nil, err
	}
	var res []map[string]string
	if err = json.Unmarshal(resp.Data, &res); err != nil {
		return nil, err
	}
	infoList := make([]*entity.ServerInfo, 0, len(res))
	for _, d := range res {
		infoList = append(infoList, &entity.ServerInfo{
			ServerID: d["serverId"],
			HttpAddr: net.JoinHostPort(d["location"], d["httpPort"]),
			RpcAddr:  net.JoinHostPort(d["location"], d["grpcPort"]),
		})
	}
	return infoList, nil
}

func (*Metadata) GetConfig(ip string) ([]byte, error) {
	cc, err := grpc.Dial(ip, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	client := pb.NewConfigServiceClient(cc)
	resp, err := client.GetConfig(context.Background(), new(pb.EmptyReq))
	if err != nil {
		return nil, response.NewError(400, err.Error())
	}
	return resp.JsonEncode, nil
}
