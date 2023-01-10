package logic

import (
	"adminserver/internal/usecase/pool"
	"adminserver/internal/usecase/webapi"
	"common/pb"
	"common/response"
	"common/util"
	"common/util/crypto"
	"context"
	"google.golang.org/grpc"
	"io"
	"mime/multipart"
)

type Objects struct {
}

func NewObjects() *Objects {
	return &Objects{}
}

func (Objects) Upload(file *multipart.FileHeader, token string) error {
	// open and checksum
	temp, err := file.Open()
	if err != nil {
		return err
	}
	hash := crypto.SHA256IO(temp)
	util.LogErr(temp.Close())
	// open and send request
	fileBody, err := file.Open()
	if err != nil {
		return err
	}
	return webapi.PutObjects(SelectApiServer(), file.Filename, hash, fileBody, file.Size, token)
}

func (Objects) Download(name string, version int, token string) (io.ReadCloser, error) {
	return webapi.GetObjects(SelectApiServer(), name, version, token)
}

func (Objects) JoinCluster(serverId string) error {
	mp := pool.Discovery.GetServiceMapping(pool.Config.Discovery.DataServName, true)
	addr, ok := mp[serverId]
	if !ok {
		return response.NewError(400, "serverId not exist")
	}
	cc, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	client := pb.NewObjectMigrationClient(cc)
	resp, err := client.JoinCommand(context.Background(), new(pb.EmptyReq))
	if err != nil {
		return err
	}
	if !resp.Success {
		return response.NewError(400, resp.Message)
	}
	return nil
}

func (Objects) LeaveCluster(serverId string) error {
	mp := pool.Discovery.GetServiceMapping(pool.Config.Discovery.DataServName, true)
	addr, ok := mp[serverId]
	if !ok {
		return response.NewError(400, "serverId not exist")
	}
	cc, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	client := pb.NewObjectMigrationClient(cc)
	resp, err := client.LeaveCommand(context.Background(), new(pb.EmptyReq))
	if err != nil {
		return err
	}
	if !resp.Success {
		return response.NewError(400, resp.Message)
	}
	return nil
}
