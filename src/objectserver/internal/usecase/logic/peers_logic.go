package logic

import (
	"objectserver/internal/usecase/pool"
)

type Peers struct{}

func NewPeers() Peers {
	return Peers{}
}

func (p Peers) GetPeerMap() map[string]string {
	return pool.Discovery.GetServiceMapping(pool.Config.Discovery.MetaServName, true)
}
