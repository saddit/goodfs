package grpc

import (
	"common/pb"
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"metaserver/internal/usecase"
)

type MetadataApiServer struct {
	pb.UnimplementedMetadataApiServer
	Service usecase.IMetadataService
}

func NewMetadataApiServer(s usecase.IMetadataService) *MetadataApiServer {
	return &MetadataApiServer{Service: s}
}

func (m *MetadataApiServer) GetVersionsByHash(_ context.Context, req *pb.ApiQryHash) (*pb.ApiQryResp, error) {
	if req == nil || req.Hash == "" {
		return nil, status.Error(codes.InvalidArgument, "hash value required")
	}
	res, err := m.Service.FindByHash(req.Hash)
	if err != nil {
		return nil, err
	}
	return &pb.ApiQryResp{Data: res}, nil
}
