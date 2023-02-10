package grpc

import (
	"common/proto/msg"
	"common/proto/pb"
	"common/response"
	"common/util"
	"context"
	"github.com/gin-gonic/gin/binding"
	"github.com/tinylib/msgp/msgp"
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

var emp = new(pb.Empty)

func NewMetadataApiServer(s usecase.IMetadataService, b usecase.BucketService) *MetadataApiServer {
	return &MetadataApiServer{Service: s, BucketService: b}
}

func (m *MetadataApiServer) GetVersionsByHash(_ context.Context, req *pb.MetaReq) (*pb.Msgpack, error) {
	if req.Hash == "" {
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

func (m *MetadataApiServer) ListVersion(_ context.Context, req *pb.MetaReq) (*pb.Msgpack, error) {
	if req.Id == "" || req.Page == nil {
		return nil, status.Error(codes.InvalidArgument, "id and page required")
	}
	if req.Page.Page <= 0 || req.Page.PageSize <= 0 {
		return nil, status.Error(codes.InvalidArgument, "page and pageSize must gt 0")
	}
	res, total, err := m.Service.ListVersions(req.Id, int(req.Page.Page), int(req.Page.PageSize))
	if err != nil {
		return nil, err
	}
	bt, err := util.EncodeArrayMsgp(res)
	if err != nil {
		return nil, response.GRPCError(err)
	}
	return &pb.Msgpack{Data: bt, Total: int64(total)}, nil
}

func (m *MetadataApiServer) GetBucket(_ context.Context, req *pb.MetaReq) (*pb.Msgpack, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id (bucket name) required")
	}
	bt, err := m.BucketService.GetBytes(req.Id)
	if err != nil {
		return nil, response.GRPCError(err)
	}
	return &pb.Msgpack{Data: bt}, nil
}
func (m *MetadataApiServer) GetMetadata(_ context.Context, req *pb.MetaReq) (*pb.Msgpack, error) {
	if req.Id == "" {
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
	if req.Id == "" {
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

func (m *MetadataApiServer) SaveMetadata(_ context.Context, req *pb.Metadata) (*pb.Empty, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "metadata id required")
	}
	var md msg.Metadata
	if err := ShouldBindMsgpack(&md, req.Msgpack); err != nil {
		return nil, response.GRPCError(err)
	}
	if err := m.Service.AddMetadata(req.Id, &md); err != nil {
		return nil, response.GRPCError(err)
	}
	return emp, nil
}

func (m *MetadataApiServer) SaveBucket(_ context.Context, req *pb.Metadata) (*pb.Empty, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "metadata id required")
	}
	var md msg.Bucket
	if err := ShouldBindMsgpack(&md, req.Msgpack); err != nil {
		return nil, response.GRPCError(err)
	}
	md.Name = req.Id
	if err := m.BucketService.Create(&md); err != nil {
		return nil, response.GRPCError(err)
	}
	return emp, nil
}

func (m *MetadataApiServer) SaveVersion(_ context.Context, req *pb.Metadata) (*pb.Int32, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "metadata id required")
	}
	var md msg.Version
	if err := ShouldBindMsgpack(&md, req.Msgpack); err != nil {
		return nil, response.GRPCError(err)
	}
	vn, err := m.Service.AddVersion(req.Id, &md)
	if err != nil {
		return nil, response.GRPCError(err)
	}
	return &pb.Int32{Data: int32(vn)}, nil
}

func (m *MetadataApiServer) UpdateVersion(_ context.Context, req *pb.Metadata) (*pb.Empty, error) {
	if req.Id == "" || req.Version <= 0 {
		return nil, status.Error(codes.InvalidArgument, "metadata id and version required")
	}
	var md msg.Version
	if err := ShouldBindMsgpack(&md, req.Msgpack); err != nil {
		return nil, response.GRPCError(err)
	}
	if err := m.Service.UpdateVersion(req.Id, int(req.Version), &md); err != nil {
		return nil, response.GRPCError(err)
	}
	return emp, nil
}

func (m *MetadataApiServer) GetPeers(context.Context, *pb.Empty) (*pb.Strings, error) {
	peers, err := logic.NewPeers().GetPeers()
	if err != nil {
		return nil, response.GRPCError(err)
	}
	if err != nil {
		return nil, response.GRPCError(err)
	}
	return &pb.Strings{Data: peers}, nil
}

func ShouldBindMsgpack(data msgp.Unmarshaler, bt []byte) error {
	if err := util.DecodeMsgp(data, bt); err != nil {
		return status.Error(codes.InvalidArgument, "invalid request value")
	}
	if err := binding.Validator.ValidateStruct(data); err != nil {
		return response.GRPCError(err)
	}
	return nil
}
