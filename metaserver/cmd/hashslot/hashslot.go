package hashslot

import (
	"common/cmd"
	"common/util"
	"context"
	"fmt"
	"metaserver/internal/usecase/pb"
	"strings"

	"google.golang.org/grpc"
)

var address string

func init() {
	cmd.Register("start-migration", func(args []string) {
		if len(args) != 4 {
			fmt.Println("start-migration rpc-port target-host target-rpc-port a-b,c-d")
			return
		}
		address = util.GetHostPort(address)
		startMigration(args[1], args[2], strings.Split(args[3], ","))
	})
}

func getClient() (pb.HashSlotClient, error) {
	cc, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return pb.NewHashSlotClient(cc), nil
}

func startMigration(targetHost, rpcPort string, slots []string) {
	cli, err := getClient()
	if err != nil {
		return
	}
	resp, err := cli.StartMigration(context.Background(), &pb.MigrationReq{
		Slots: slots,
		TargetLocation: &pb.LocationInfo{
			Host: targetHost,
			RpcPort: rpcPort,
		},
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("success: %v, message: %s\n", resp.Success, resp.Message)
}