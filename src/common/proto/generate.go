package proto

import (
	"common/proto/pb"
	"common/response"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//sudo apt install -y protobuf-compiler
//go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
//go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
//go:generate protoc -I=. --go_out=. --go-grpc_out=. raft_cmd.proto
//go:generate protoc -I=. --go_out=. --go-grpc_out=. message.proto
//go:generate protoc -I=. --go_out=. --go-grpc_out=. hashslot.proto
//go:generate protoc -I=. --go_out=. --go-grpc_out=. metadata.proto
//go:generate protoc -I=. --go_out=. --go-grpc_out=. object_migration.proto
//go:generate protoc -I=. --go_out=. --go-grpc_out=. config_service.proto

func ResolveErr(err error) error {
	s, ok := status.FromError(err)
	if !ok {
		return response.NewError(500, err.Error())
	}
	switch s.Code() {
	case codes.OK:
		return nil
	case codes.NotFound:
		return response.NewError(404, s.Message())
	case codes.InvalidArgument, codes.Aborted:
		return response.NewError(400, s.Message())
	default:
		return response.NewError(500, s.Message())
	}
}

func ResolveResponse(resp *pb.Response, err error) (string, error) {
	if err != nil {
		return "", err
	}
	if resp.Success {
		return resp.Message, nil
	}
	return "", errors.New(resp.Message)
}
