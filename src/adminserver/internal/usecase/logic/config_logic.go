package logic

import (
	"common/proto/pb"
	"common/response"
	"context"
)

type ConfigLogic struct {
}

func (ConfigLogic) GetConfig(ip string) ([]byte, error) {
	cc, err := getConn(ip)
	if err != nil {
		return nil, err
	}
	client := pb.NewConfigServiceClient(cc)
	resp, err := client.GetConfig(context.Background(), new(pb.EmptyReq))
	if err != nil {
		return nil, response.NewError(400, err.Error())
	}
	return resp.YamlEncode, nil
}
