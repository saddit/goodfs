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
	"net"
	"sort"
	"sync"

	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type MetadataCond struct {
	Name     string `form:"name"`
	Version  int    `form:"version"`
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
			data, total, err := webapi.ListMetadata(loc, cond.Name, cond.Page*cond.PageSize, cond.OrderBy, cond.Desc)
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

func (m *Metadata) VersionPaging(cond MetadataCond) ([]byte, int, error) {
	return webapi.ListVersion(SelectApiServer(), cond.Name, cond.Page, cond.PageSize)
}

func (m *Metadata) StartMigration(srcID, destID string, slots []string) error {
	mp := pool.Discovery.GetServiceMapping(pool.Config.Discovery.MetaServName, true)
	srcAddr, destAddr := mp[srcID], mp[destID]
	if srcAddr == "" || destAddr == "" {
		return response.NewError(400, "invalid server id")
	}
	cc, err := grpc.Dial(srcAddr)
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
		if err := util.DecodeMsgp(&info, kv.Value); err != nil {
			return nil, err
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
	for _, mp := range res {
		infoList = append(infoList, &entity.ServerInfo{
			ServerID: mp["serverId"],
			HttpAddr: net.JoinHostPort(mp["location"], mp["httpPort"]),
			RpcAddr:  net.JoinHostPort(mp["location"], mp["grpcPort"]),
		})
	}
	return infoList, nil
}
