package grpc

import (
	"common/collection/set"
	"common/graceful"
	"common/proto/pb"
	"common/util"
	"common/util/slices"
	"context"
	"encoding/json"
	"fmt"
	"metaserver/config"
	"metaserver/internal/usecase/logic"
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

func (rcs *RaftCmdServerImpl) Bootstrap(_ context.Context, req *pb.BootstrapReq) (*pb.Response, error) {
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
		if len(pool.Config.Cluster.Nodes) == 0 {
			pool.Config.Cluster.Nodes = append(pool.Config.Cluster.Nodes, fmt.Sprint(pool.Config.Cluster, ",", util.ServerAddress(pool.Config.RpcPort)))
		}
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

func (rcs *RaftCmdServerImpl) AddVoter(_ context.Context, req *pb.AddVoterReq) (*pb.Response, error) {
	if len(req.GetVoters()) == 0 {
		return &pb.Response{Success: false, Message: "empty voters"}, nil
	}
	var resultFeatures []raft.IndexFuture
	for _, item := range req.GetVoters() {
		res := rcs.rf.Raft.AddVoter(raft.ServerID(item.GetId()), raft.ServerAddress(item.GetAddress()), item.GetPrevIndex(), 10*time.Second)
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
	// save to config
	nodes := set.OfString(pool.Config.Cluster.Nodes)
	for _, item := range req.GetVoters() {
		node := fmt.Sprint(item.Id, ",", item.Address)
		if nodes.Contains(node) {
			continue
		}
		pool.Config.Cluster.Nodes = append(pool.Config.Cluster.Nodes)
	}
	// persist config async
	go func() {
		defer graceful.Recover()
		util.LogErrWithPre("persist config", pool.Config.Persist())
	}()
	return okResp, nil
}

func (rcs *RaftCmdServerImpl) Config(context.Context, *pb.EmptyReq) (*pb.Response, error) {
	bt, err := json.Marshal(pool.Config.Cluster)
	if err != nil {
		return nil, err
	}
	return &pb.Response{Success: true, Message: util.BytesToStr(bt)}, nil
}

func (rcs *RaftCmdServerImpl) LeaveCluster(ctx context.Context, _ *pb.EmptyReq) (*pb.Response, error) {
	if rcs.rf.IsLeader() {
		if err := rcs.rf.Raft.LeadershipTransfer().Error(); err != nil {
			return nil, fmt.Errorf("transfer leader error: %w", err)
		}
		return okResp, nil
	}
	leaderAddr := rcs.rf.LeaderAddress()
	cc, err := grpc.Dial(leaderAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return pb.NewRaftCmdClient(cc).RemoveFollower(ctx, &pb.RemoveFollowerReq{
		FollowerId: rcs.rf.ID,
		PrevIndex:  rcs.rf.Raft.AppliedIndex(),
	})
}

func (rcs *RaftCmdServerImpl) RemoveFollower(_ context.Context, req *pb.RemoveFollowerReq) (*pb.Response, error) {
	feature := rcs.rf.Raft.DemoteVoter(raft.ServerID(req.FollowerId), req.PrevIndex, time.Second)
	if err := feature.Error(); err != nil {
		return nil, err
	}
	id := fmt.Sprint(req.FollowerId, ",")
	idx := -1
	for i, node := range pool.Config.Cluster.Nodes {
		if strings.HasPrefix(node, id) {
			idx = i
			break
		}
	}
	if idx > 0 {
		pool.Config.Cluster.Nodes[0], pool.Config.Cluster.Nodes[idx] = pool.Config.Cluster.Nodes[idx], pool.Config.Cluster.Nodes[idx]
		slices.RemoveFirst(&pool.Config.Cluster.Nodes)
		// persist config async
		go func() {
			defer graceful.Recover()
			util.LogErrWithPre("persist config", pool.Config.Persist())
		}()
	}
	return okResp, nil
}

func (rcs *RaftCmdServerImpl) JoinLeader(ctx context.Context, req *pb.JoinLeaderReq) (*pb.Response, error) {
	// dial a connection to leader
	mp := pool.Registry.GetServiceMapping(pool.Config.Registry.Name, true)
	addr, ok := mp[req.MasterId]
	if !ok {
		return nil, fmt.Errorf("'%s' is not an exist meta-server id", req.MasterId)
	}
	cc, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	// send rpc request by client
	client := pb.NewRaftCmdClient(cc)
	if rcs.rf.LeaderID() != req.MasterId {
		if rcs.rf.LeaderID() != "" {
			// demote voter
			_, err := client.RemoveFollower(ctx, &pb.RemoveFollowerReq{
				FollowerId: rcs.rf.ID,
				PrevIndex:  rcs.rf.Raft.AppliedIndex(),
			})
			if err != nil {
				return nil, err
			}
		}
		resp, err := client.AddVoter(ctx, &pb.AddVoterReq{
			Voters: []*pb.Voter{{
				Id:        rcs.rf.ID,
				Address:   rcs.rf.Address,
				PrevIndex: rcs.rf.Raft.AppliedIndex(),
			}},
		})
		if err != nil {
			return nil, err
		}
		if !resp.Success {
			return resp, nil
		}
	}
	// get leader config
	resp, err := client.Config(context.Background(), new(pb.EmptyReq))
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		return resp, nil
	}
	var cfg config.ClusterConfig
	if err := json.Unmarshal(util.StrToBytes(resp.Message), &cfg); err != nil {
		return nil, err
	}
	err = logic.NewRaftCluster().UpdateConfiguration(&cfg)
	return okResp, err
}

func (rcs *RaftCmdServerImpl) AppliedIndex(_ context.Context, _ *pb.EmptyReq) (*pb.Response, error) {
	return &pb.Response{Success: true, Message: fmt.Sprint(rcs.rf.Raft.AppliedIndex())}, nil
}

func (rcs *RaftCmdServerImpl) Peers(_ context.Context, _ *pb.EmptyReq) (*pb.Response, error) {
	servers := rcs.rf.Raft.GetConfiguration().Configuration().Servers
	res := make([]string, 0, len(servers)-1)
	for _, serv := range servers {
		res = append(res, string(serv.Address))
	}
	return &pb.Response{Success: true, Message: strings.Join(res, ",")}, nil
}
