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

func (m *Metadata) MetadataPaging(cond MetadataCond) ([]*entity.Metadata, error) {
	servers := pool.Discovery.GetServices(pool.Config.Discovery.MetaServName, false)
	lst := make([]*entity.Metadata, 0, len(servers)*cond.Page*cond.PageSize)
	if len(servers) == 0 {
		logs.Std().Warn("not found any metadata server")
		return lst, nil
	}
	mux := sync.Mutex{}
	dg := util.NewDoneGroup()
	defer dg.Close()
	for _, ip := range servers {
		dg.Add(1)
		go func(loc string) {
			defer dg.Done()
			data, err := webapi.ListMetadata(loc, cond.Name, cond.Page*cond.PageSize, cond.OrderBy, cond.Desc)
			if err != nil {
				dg.Error(err)
				return
			}
			mux.Lock()
			defer mux.Unlock()
			lst = append(lst, data...)
		}(ip)
	}
	if err := dg.WaitUntilError(); err != nil {
		return nil, err
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
		return lst[st:ed], nil
	}
	return []*entity.Metadata{}, nil
}

func (m *Metadata) VersionPaging(cond MetadataCond) ([]byte, error) {
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

func (m *Metadata) JoinRaftCluster(masterId, servId string) error {
	mp := pool.Discovery.GetServiceMapping(pool.Config.Discovery.MetaServName, true)
	masterAddr, ok := mp[masterId]
	if !ok {
		return response.NewError(400, "unknown master id")
	}
	servAddr, ok := mp[servId]
	if !ok {
		return response.NewError(400, "unknown server id")
	}
	cc, err := grpc.Dial(servAddr)
	if err != nil {
		return err
	}
	client := pb.NewRaftCmdClient(cc)
	resp, err := client.JoinLeader(context.Background(), &pb.JoinLeaderReq{Address: masterAddr})
	if err != nil {
		return err
	}
	if !resp.Success {
		return response.NewError(400, resp.Message)
	}
	return nil
}
