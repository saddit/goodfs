package logic

import (
	"adminserver/internal/entity"
	"adminserver/internal/usecase/pool"
	"adminserver/internal/usecase/webapi"
	"common/collection/set"
	"common/cst"
	"common/hashslot"
	"common/logs"
	"common/proto/msg"
	"common/proto/pb"
	"common/response"
	"common/util"
	"fmt"
	"net"
	"sort"
	"sync"

	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/net/context"
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

func NewMetadata() Metadata {
	return Metadata{}
}

func (m Metadata) MetadataPaging(cond *MetadataCond) ([]*msg.Metadata, int, error) {
	servers := pool.Discovery.GetServicesWith(pool.Config.Discovery.MetaServName, true)
	lst := make([]*msg.Metadata, 0, len(servers)*cond.Page*cond.PageSize)
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
	sort.Slice(lst, func(i, j int) bool {
		return lst[i].ID() < lst[j].ID()
	})
	if st, ed, ok := util.PagingOffset(cond.Page, cond.PageSize, len(lst)); ok {
		return lst[st:ed], totals, nil
	}
	return []*msg.Metadata{}, 0, nil
}

func (m Metadata) VersionPaging(cond MetadataCond, token string) ([]byte, int, error) {
	return webapi.ListVersion(SelectApiServer(), cond.Name, cond.Bucket, cond.Page, cond.PageSize, token)
}

func (m Metadata) BucketPaging(cond *BucketCond) ([]*msg.Bucket, int, error) {
	servers := pool.Discovery.GetServicesWith(pool.Config.Discovery.MetaServName, true)
	lst := make([]*msg.Bucket, 0, len(servers)*cond.Page*cond.PageSize)
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
	sort.Slice(lst, func(i, j int) bool {
		return lst[i].ID() < lst[j].ID()
	})
	st, ed, _ := util.PagingOffset(cond.Page, cond.PageSize, len(lst))
	return lst[st:ed], totals, nil
}

func (m Metadata) StartMigration(srcID, destID string, slots []string) error {
	mp := pool.Discovery.GetServiceMapping(pool.Config.Discovery.MetaServName)
	srcAddr, destAddr := mp[srcID], mp[destID]
	if srcAddr == "" || destAddr == "" {
		return response.NewError(400, "invalid server id")
	}
	cc, err := getConn(srcAddr)
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
			Host: host,
			Port: port,
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

func (m Metadata) GetSlotsDetail() (map[string]*hashslot.SlotInfo, error) {
	prefix := cst.EtcdPrefix.FmtHashSlot(pool.Config.Discovery.Group, "")
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

func (m Metadata) GetMasterServerIds() set.Set {
	mp := pool.Discovery.GetServiceMappingWith(pool.Config.Discovery.MetaServName, true)
	masters := set.OfMapKeys(mp)
	return masters
}

func (m Metadata) LeaveRaftCluster(servId string) error {
	mp := pool.Discovery.GetServiceMapping(pool.Config.Discovery.MetaServName)
	servAddr, ok := mp[servId]
	if !ok {
		return response.NewError(400, "unknown server id")
	}
	cc, err := getConn(servAddr)
	if err != nil {
		return err
	}
	_, err = pb.NewRaftCmdClient(cc).LeaveCluster(context.Background(), new(pb.EmptyReq))
	return err
}

func (m Metadata) JoinRaftCluster(masterId, servId string) error {
	mp := pool.Discovery.GetServiceMapping(pool.Config.Discovery.MetaServName)
	servAddr, ok := mp[servId]
	if !ok {
		return response.NewError(400, "unknown server id")
	}
	cc, err := getConn(servAddr)
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

func (m Metadata) GetPeers(servId string) ([]*entity.ServerInfo, error) {
	ipMap := pool.Discovery.GetServiceMapping(pool.Config.Discovery.MetaServName)
	servAddr, ok := ipMap[servId]
	if !ok {
		return nil, response.NewError(400, "unknown server id")
	}
	cc, err := getConn(servAddr)
	if err != nil {
		return nil, err
	}
	resp, err := pb.NewMetadataApiClient(cc).GetPeers(context.Background(), new(pb.Empty))
	if err != nil {
		return nil, ResolveErr(err)
	}
	resp.Data = append(resp.Data, servId)
	var res []map[string]string
	infoList := make([]*entity.ServerInfo, 0, len(res))
	for _, id := range resp.Data {
		infoList = append(infoList, &entity.ServerInfo{
			ServerID: id,
			HttpAddr: ipMap[id],
		})
	}
	return infoList, nil
}

func (m Metadata) TotalBuckets() (int, error) {
	_, total, err := m.BucketPaging(&BucketCond{
		Page:     1,
		PageSize: 1,
	})
	return total, err
}

func (m Metadata) TotalObjects() (int, error) {
	_, total, err := m.MetadataPaging(&MetadataCond{
		Page:     1,
		PageSize: 1,
	})
	return total, err
}
