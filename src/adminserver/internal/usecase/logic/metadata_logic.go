package logic

import (
	"adminserver/internal/entity"
	"adminserver/internal/usecase/pool"
	"adminserver/internal/usecase/webapi"
	"common/util"
	"math/rand"
	"sort"
	"sync"
	"time"
)

type MetadataCond struct {
	Name     string `form:"name"`
	Version  int    `form:"version"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	OrderBy  string `form:"order_by"`
	Desc     bool   `form:"desc"`
}

type Metadata struct{}

func NewMetadata() *Metadata {
	return new(Metadata)
}

func (m *Metadata) MetadataPaging(cond MetadataCond) ([]*entity.Metadata, error) {
	servers := pool.Discovery.GetServices(pool.Config.Discovery.MetaServName, false)
	lst := make([]*entity.Metadata, 0, len(servers)*cond.Page*cond.PageSize)
	dg := util.NewDoneGroup()
	mux := sync.Mutex{}
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
	servers := pool.Discovery.GetServices(pool.Config.Discovery.ApiServName, false)
	rand.Seed(time.Now().Unix())
	idx := rand.Intn(len(servers))
	ip := servers[idx]
	return webapi.ListVersion(ip, cond.Name, cond.Page, cond.PageSize)
}
