package grpc

import (
	"common/proto/pb"
	"context"
	"gopkg.in/yaml.v3"
	"metaserver/internal/usecase/pool"
)

type ConfigServiceServer struct {
	pb.UnimplementedConfigServiceServer
}

func (o *ConfigServiceServer) GetConfig(context.Context, *pb.EmptyReq) (*pb.ConfigResp, error) {
	conf := *pool.Config
	conf.Etcd.Username = "*****"
	conf.Etcd.Password = "*****"
	bt, err := yaml.Marshal(&conf)
	if err != nil {
		return nil, err
	}
	return &pb.ConfigResp{YamlEncode: bt}, nil
}
