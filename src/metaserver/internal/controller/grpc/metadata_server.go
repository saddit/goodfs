package grpc

import (
	"common/pb"
	"context"
	"encoding/json"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"metaserver/internal/usecase"
	"metaserver/internal/usecase/logic"
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

func (m *MetadataApiServer) GetPeers(_ context.Context, _ *pb.EmptyReq) (*pb.JsonResp, error) {
	peers, err := logic.NewPeers().GetPeers()
	if err != nil {
		return nil, err
	}
	bt, err := json.Marshal(peers)
	if err != nil {
		return nil, err
	}
	return &pb.JsonResp{Data: bt}, nil
}
