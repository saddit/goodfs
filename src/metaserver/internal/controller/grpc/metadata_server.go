package grpc

import (
	"common/proto/pb"
	"common/response"
	"common/util"
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"metaserver/internal/usecase"
	"metaserver/internal/usecase/logic"
)

type MetadataApiServer struct {
	pb.UnimplementedMetadataApiServer
	Service       usecase.IMetadataService
	BucketService usecase.BucketService
}

func NewMetadataApiServer(s usecase.IMetadataService, b usecase.BucketService) *MetadataApiServer {
	return &MetadataApiServer{Service: s, BucketService: b}
}

func (m *MetadataApiServer) GetVersionsByHash(_ context.Context, req *pb.MetaReq) (*pb.Msgpack, error) {
	if req == nil || req.Hash == "" {
		return nil, status.Error(codes.InvalidArgument, "hash value required")
	}
	res, err := m.Service.FindByHash(req.Hash)
	if err != nil {
		return nil, err
	}
	bt, err := util.EncodeArrayMsgp(res)
	if err != nil {
		return nil, response.GRPCError(err)
	}
	return &pb.Msgpack{Data: bt}, nil
}

func (m *MetadataApiServer) GetBucket(_ context.Context, req *pb.MetaReq) (*pb.Msgpack, error) {
	if req == nil || req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id (bucket name) required")
	}
	bt, err := m.BucketService.GetBytes(req.Id)
	if err != nil {
		return nil, response.GRPCError(err)
	}
	return &pb.Msgpack{Data: bt}, nil
}
func (m *MetadataApiServer) GetMetadata(_ context.Context, req *pb.MetaReq) (*pb.Msgpack, error) {
	if req == nil || req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "metadata id required")
	}
	md, _, err := m.Service.GetMetadata(req.Id, -1, req.WithExtra)
	if err != nil {
		return nil, response.GRPCError(err)
	}
	bt, err := util.EncodeMsgp(md)
	if err != nil {
		return nil, response.GRPCError(err)
	}
	return &pb.Msgpack{Data: bt}, nil
}
func (m *MetadataApiServer) GetVersion(_ context.Context, req *pb.MetaReq) (*pb.Msgpack, error) {
	if req == nil || req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "metadata id required")
	}
	v, err := m.Service.GetVersion(req.Id, int(req.Version))
	if err != nil {
		return nil, response.GRPCError(err)
	}
	bt, err := util.EncodeMsgp(v)
	if err != nil {
		return nil, response.GRPCError(err)
	}
	return &pb.Msgpack{Data: bt}, nil
}

func (m *MetadataApiServer) GetPeers(_ context.Context, _ *pb.EmptyReq) (*pb.StringsResp, error) {
	peers, err := logic.NewPeers().GetPeers()
	if err != nil {
		return nil, response.GRPCError(err)
	}
	if err != nil {
		return nil, response.GRPCError(err)
	}
	return &pb.StringsResp{Data: peers}, nil
}
