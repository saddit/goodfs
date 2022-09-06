package grpc

import (
	"context"
	"fmt"
	"metaserver/internal/usecase/pb"
	"metaserver/internal/usecase/pool"
	"metaserver/internal/usecase/raftimpl"
	"strings"
	"time"

	"github.com/hashicorp/raft"
	"google.golang.org/grpc"
)

type RaftCmdServerImpl struct {
	pb.UnimplementedRaftCmdServer
	rf *raftimpl.RaftWrapper
}

var okResp = &pb.Response{Success: true, Message: "ok"}

func NewRaftCmdServer(r *raftimpl.RaftWrapper) pb.RaftCmdServer {
	return &RaftCmdServerImpl{rf: r}
}

func (rcs *RaftCmdServerImpl) Bootstrap(ctx context.Context, req *pb.BootstrapReq) (*pb.Response, error) {
	// init voter
	var raftCfg raft.Configuration
	if len(req.GetServices()) > 0 {
		raftCfg.Servers = make([]raft.Server, 0, len(req.GetServices()))
		for _, v := range req.GetServices() {
			raftCfg.Servers = append(raftCfg.Servers, raft.Server{
				Suffrage: raft.Voter,
				ID:       raft.ServerID(v.GetId()),
				Address:  raft.ServerAddress(v.GetAddress()),
			})
		}
	} else {
		raftCfg.Servers = make([]raft.Server, len(pool.Config.Cluster.Nodes))
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
	// bootstrap
	f := rcs.rf.Raft.BootstrapCluster(raftCfg)
	if err := f.Error(); err != nil {
		return &pb.Response{Success: false, Message: err.Error()}, nil
	}

	return okResp, nil
}

func (rcs *RaftCmdServerImpl) AddVoter(ctx context.Context, req *pb.AddVoterReq) (*pb.Response, error) {
	if len(req.GetVoters()) == 0 {
		return &pb.Response{Success: false, Message: "empty voters"}, nil
	}
	var resultFeatures []raft.IndexFuture
	for _, item := range req.GetVoters() {
		res := rcs.rf.Raft.AddVoter(raft.ServerID(item.GetId()), raft.ServerAddress(item.GetAddress()), item.GetPrevIndex(), 10 * time.Second)
		resultFeatures = append(resultFeatures, res)
	}
	var msg strings.Builder
	for i, feature := range resultFeatures {
		if err := feature.Error(); err != nil {
			voter := req.GetVoters()[i]
			msg.WriteString(fmt.Sprintf("%s-%s add fail: %s;", voter.GetId(), voter.GetAddress(), err.Error()))
		}
	}
	if msg.Len() > 0 {
		return &pb.Response{Success: false, Message: msg.String()}, nil
	}
	return okResp, nil
}

func (rcs *RaftCmdServerImpl) JoinLeader(ctx context.Context, req *pb.JoinLeaderReq) (*pb.Response, error) {
	// dial a connection to leader
	cc, err := grpc.Dial(req.GetAddress())
	if err != nil {
		return nil, err
	}
	// send rpc request by client
	client := pb.NewRaftCmdClient(cc)
	return client.AddVoter(ctx, &pb.AddVoterReq{
		Voters: []*pb.Voter{{
			Id: rcs.rf.ID,
			Address: rcs.rf.Address,
			PrevIndex: rcs.rf.Raft.AppliedIndex(),
		}},
	})
}

func (rcs *RaftCmdServerImpl) AppliedIndex(_ context.Context, _ *pb.EmptyReq) (*pb.Response, error) {
	return &pb.Response{Success: true, Message: fmt.Sprint(rcs.rf.Raft.AppliedIndex())}, nil
}