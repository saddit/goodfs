package grpc

import (
	"common/proto/pb"
	"context"
	"encoding/json"
	"metaserver/internal/usecase/pool"
)

type ConfigServiceServer struct {
	pb.UnimplementedConfigServiceServer
}

func (o *ConfigServiceServer) GetConfig(context.Context, *pb.EmptyReq) (*pb.ConfigResp, error) {
	conf := pool.Config
	bt, err := json.Marshal(&conf)
	if err != nil {
		return nil, err
	}
	return &pb.ConfigResp{JsonEncode: bt}, nil
}
