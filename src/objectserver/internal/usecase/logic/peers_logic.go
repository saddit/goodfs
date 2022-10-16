package logic

import (
	. "common/constrant"
	"common/util"
	"context"
	"objectserver/internal/entity"
	"objectserver/internal/usecase/pool"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type Peers struct{}

func NewPeers() Peers {
	return Peers{}
}

func (Peers) GetPeers() ([]*entity.PeerInfo, error) {
	prefix := EtcdPrefix.FmtPeersInfo(pool.Config.Registry.Name, "")
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

func (p Peers) GetPeerMap() (map[string]*entity.PeerInfo, error) {
	lst, err := p.GetPeers()
	if err != nil {
		return nil, err
	}
	res := make(map[string]*entity.PeerInfo)
	for _, info := range lst {
		res[info.ServerID] = info
	}
	return res, nil
}

func (Peers) RegisterSelf() error {
	key := EtcdPrefix.FmtPeersInfo(pool.Config.Registry.Name, pool.Config.Registry.ServerID)
	info := &entity.PeerInfo{
		ServerID: pool.Config.Registry.ServerID,
		Location: util.GetHost(),
		HttpPort: pool.Config.Port,
		RpcPort:  pool.Config.RpcPort,
	}
	bt, err := util.EncodeMsgp(info)
	if err != nil {
		return err
	}
	_, err = pool.Etcd.Put(context.Background(), key, string(bt))
	return err
}

func (Peers) UnregisterSelf() error {
	key := EtcdPrefix.FmtPeersInfo(pool.Config.Registry.Name, pool.Config.Registry.ServerID)
	_, err := pool.Etcd.Delete(context.Background(), key)
	return err
}
