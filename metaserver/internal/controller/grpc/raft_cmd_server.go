package grpc

import (
	"context"
	"metaserver/internal/usecase/pb"
	"metaserver/internal/usecase/pool"
	"metaserver/internal/usecase/raftimpl"
	"strings"

	"github.com/hashicorp/raft"
)

type RaftCmdServerImpl struct {
	pb.UnimplementedRaftCmdServer
	rf *raftimpl.RaftWrapper
}

func NewRaftCmdServer(r *raftimpl.RaftWrapper) pb.RaftCmdServer {
	return &RaftCmdServerImpl{rf: r}
}

func (rcs *RaftCmdServerImpl) checkEnabled() *pb.Response {
	if !rcs.rf.Enabled {
		return &pb.Response{Success: false, Message: "raft is not enabled"}
	}
	return nil
}

func (rcs *RaftCmdServerImpl) Bootstrap(ctx context.Context, req *pb.BootstrapReq) (*pb.Response, error) {
	if res := rcs.checkEnabled(); res != nil {
		return res, nil
	}
	// init voter
	var raftCfg raft.Configuration
	if len(req.Services) > 0 {
		raftCfg.Servers = make([]raft.Server, 0, len(req.Services))
		for _, v := range req.Services {
			raftCfg.Servers = append(raftCfg.Servers, raft.Server{
				Suffrage: raft.Voter,
				ID:       raft.ServerID(v.Id),
				Address:  raft.ServerAddress(v.Address),
			})
		}
	} else {
		raftCfg.Servers = make([]raft.Server, len(req.Services))
		for i, v := range pool.Config.Cluster.Nodes {
			idAndAddr := strings.Split(v, ",")
			if len(idAndAddr) != 2 {
				log.Warnf("raft-bootstrap: skip node %s item doesn't support format 'id,host:port'", v)
				continue
			}
			raftCfg.Servers[i] = raft.Server{
				Suffrage: raft.Voter,
				ID:       raft.ServerID(idAndAddr[0]),
				Address:  raft.ServerAddress(idAndAddr[1]),
			}
		}
	}
	// bootsrap
	f := rcs.rf.Raft.BootstrapCluster(raftCfg)
	if err := f.Error(); err != nil {
		return &pb.Response{Success: false, Message: err.Error()}, nil
	}

	return &pb.Response{Success: true, Message: ""}, nil
}
