package grpc

import (
	"common/pb"
	"context"
	"metaserver/internal/usecase"
)

type MetadataApiServer struct {
	pb.UnimplementedMetadataApiServer
	Service usecase.IMetadataService
}

func NewMetadataApiServer(s usecase.IMetadataService) *MetadataApiServer {
	return &MetadataApiServer{Service: s}
}

func (m *MetadataApiServer) GetMetadataByHash(_ context.Context, req *pb.ApiQryHash) (*pb.ApiQryResp, error) {
	//TODO(feat): querying metadata by hash api
	panic("implement me")
}
