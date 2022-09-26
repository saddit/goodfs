package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"metaserver/internal/usecase"
	"metaserver/internal/usecase/pb"
)

type HashSlotServer struct {
	pb.UnimplementedHashSlotServer
	Service usecase.IHashSlotService
}

func NewHashSlotServer(serv usecase.IHashSlotService) *HashSlotServer {
	return &HashSlotServer{pb.UnimplementedHashSlotServer{}, serv}
}

func (h *HashSlotServer) PrepareMigration(_ context.Context, req *pb.PrepareReq) (*pb.Response, error) {
	if err := h.Service.PrepareMigrationFrom(req.GetLocation(), req.GetSlots()); err != nil {
		return &pb.Response{Success: false, Message: err.Error()}, nil
	}
	return okResp, nil
}

func (h *HashSlotServer) StartMigration(_ context.Context, req *pb.MigrationReq) (*pb.Response, error) {
	if err := h.Service.PrepareMigrationTo(req.GetTargetLocation(), req.GetSlots()); err != nil {
		return &pb.Response{Success: false, Message: err.Error()}, nil
	}
	if err := h.Service.AutoMigrate(req.GetTargetLocation(), req.GetSlots()); err != nil {
		return &pb.Response{Success: false, Message: err.Error()}, nil
	}
	return okResp, nil
}

func (h *HashSlotServer) StreamingReceive(stream pb.HashSlot_StreamingReceiveServer) (err error) {
	defer func() {
		if err != nil {
			if err2 := h.Service.FinishReceiveItem(false); err2 != nil {
				err = fmt.Errorf("%w: %s", err2, err)
			}
		} else {
			err = h.Service.FinishReceiveItem(true)
		}
	}()
	var resp pb.Response
	for {
		resp.Success, resp.Message = true, "ok"
		item, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if err := h.Service.ReceiveItem(item); err != nil {
			resp.Success = false
			resp.Message = err.Error()
		}
		if err := stream.Send(&resp); err != nil {
			return err
		}
	}
}

func (h *HashSlotServer) GetCurrentSlots(_ context.Context, _ *pb.EmptyReq) (*pb.Response, error) {
	mp, err := h.Service.GetCurrentSlots()
	if err != nil {
		return nil, err
	}
	s, err := json.MarshalIndent(mp, "", "  ")
	if err != nil {
		return nil, err
	}
	return &pb.Response{
		Success: true,
		Message: string(s),
	}, nil
}