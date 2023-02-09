package logic

import (
	"metaserver/internal/usecase/pool"
)

type Peers struct{}

func NewPeers() Peers {
	return Peers{}
}

// GetPeers peers server-id exclude self
func (Peers) GetPeers() ([]string, error) {
	if !pool.RaftWrapper.Enabled {
		return []string{}, nil
	}
	fu := pool.RaftWrapper.Raft.GetConfiguration()
	if err := fu.Error(); err != nil {
		return nil, err
	}
	var ids []string
	for _, sev := range fu.Configuration().Servers {
		ids = append(ids, string(sev.ID))
	}
	return ids, nil
}
