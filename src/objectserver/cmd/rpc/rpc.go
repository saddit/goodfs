package rpc

import (
	"common/cmd"
	"common/proto/pb"
	"common/util"
	"common/util/slices"
	"context"
	"fmt"
	"google.golang.org/grpc"
)

var address string
var fn = map[string]cmd.CommandFunc{
	"join-cluster":  JoinCluster,
	"leave-cluster": LeaveCluster,
}

func init() {
	cmd.Register("rpc", func(args []string) {
		if len(args) < 2 {
			fmt.Println("rpc [rpc-port] [join-cluster/leave-cluster] [...]")
			return
		}
		address = util.GetHostPort(args[0])
		if f, ok := fn[args[1]]; ok {
			f(slices.SafeChunk(args, 2, -1))
			return
		}
		fmt.Println("rpc [rpc-port] [join-cluster/leave-cluster] [...]")
	})
}

func getClient() (pb.ObjectMigrationClient, bool) {
	cc, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		fmt.Printf("connection to %s err: %s", address, err)
		return nil, false
	}
	return pb.NewObjectMigrationClient(cc), true
}

func JoinCluster(_ []string) {
	if cli, ok := getClient(); ok {
		resp, err := cli.JoinCommand(context.Background(), new(pb.EmptyReq))
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(resp)
	}
}

func LeaveCluster(_ []string) {
	if cli, ok := getClient(); ok {
		resp, err := cli.LeaveCommand(context.Background(), new(pb.EmptyReq))
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(resp)
	}
}
