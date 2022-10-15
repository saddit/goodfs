package logic

import (
	. "common/constrant"
	"common/util"
	"context"
	"metaserver/internal/entity"
	"metaserver/internal/usecase/pool"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type Peers struct{}

func NewPeers() Peers {
	return Peers{}
}

func (Peers) GetPeers() ([]*entity.PeerInfo, error) {
	prefix := EtcdPrefix.FmtPeersInfo(pool.Config.Cluster.GroupID, "")
	res, err := pool.Etcd.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	infoList := make([]*entity.PeerInfo, 0, len(res.Kvs))
	for _, kv := range res.Kvs {
		var info entity.PeerInfo
		if err = util.DecodeMsgp(&info, kv.Value); err != nil {
			return nil, err
		}
		infoList = append(infoList, &info)
	}
	return infoList, nil
}

func (Peers) Register() error {
	key := EtcdPrefix.FmtPeersInfo(pool.Config.Cluster.GroupID, pool.Config.Cluster.ID)
	info := &entity.PeerInfo{
		Location: util.GetHost(),
		HttpPort: pool.Config.Port,
		GrpcPort: pool.Config.Cluster.Port,
		GroupID:  pool.Config.Cluster.GroupID,
	}
	bt, err := util.EncodeMsgp(info)
	if err != nil {
		return err
	}
	_, err = pool.Etcd.Put(context.Background(), key, string(bt))
	return err
}

func (p Peers) MustRegister() Peers {
	util.PanicErr(p.Register())
	return p
}

func (Peers) Unregister() error {
	key := EtcdPrefix.FmtPeersInfo(pool.Config.Cluster.GroupID, pool.Config.Cluster.ID)
	_, err := pool.Etcd.Delete(context.Background(), key)
	return err
}
