package raft

import (
	"common/cmd"
	"common/util"
	"context"
	"fmt"
	"metaserver/internal/usecase/pb"
	"strings"

	"google.golang.org/grpc"
)

var port string

func init() {
	cmd.Register("raft", func(args []string) {
		if len(args) < 2 {
			fmt.Println("should input command: raft port [add_voter/join_leader/applied_index] [..]")
			return
		}
		port = args[0]
		switch args[1] {
		case "add_voter":
			if len(args) < 3 {
				fmt.Println("add_voter require 'id,host:port,index'")
				return
			}
			addVoter(args[2:])
		case "join_leader":
			if len(args) < 3 {
				fmt.Println("join_leader require leader's address")
				return
			}
			joinLeader(args[2])
		case "applied_index":
			getAppliedIndex()
		default:
			fmt.Printf("no such command %s\n", args[1])
		}
	})
}

func getClient() (pb.RaftCmdClient, error) {
	cc, err := grpc.Dial(fmt.Sprint(":", port))
	if err != nil {
		return nil, err
	}
	return pb.NewRaftCmdClient(cc), nil
}

func addVoter(voters []string) {
	cli, err := getClient() 
	if err != nil {
		return
	}
	req := &pb.AddVoterReq{Voters: make([]*pb.Voter, len(voters))}
	for i, v := range voters {
		args := strings.Split(v, ",")
		if len(args) != 3 {
			fmt.Printf("voter (%s) format error, reuqire id,host:port,index\n", v)
			continue	
		}
		req.Voters[i].Id = args[0]
		req.Voters[i].Address = args[1]
		req.Voters[i].PrevIndex = util.ToUint64(args[2])
	}
	resp, err := cli.AddVoter(context.Background(), req)
	fmt.Println(util.IfElse(err == nil, resp.Message, err.Error()))
}

func joinLeader(leader string) {
	cli, err := getClient() 
	if err != nil {
		return
	}
	resp, err := cli.JoinLeader(context.Background(), &pb.JoinLeaderReq{Address: leader})
	fmt.Println(util.IfElse(err == nil, resp.Message, err.Error()))
}

func getAppliedIndex() {
	cli, err := getClient() 
	if err != nil {
		return
	}
	resp, err := cli.AppliedIndex(context.Background(), &pb.EmptyReq{})
	fmt.Println(util.IfElse(err == nil, resp.Message, err.Error()))
}