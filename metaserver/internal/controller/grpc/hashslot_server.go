package grpc

import (
	"metaserver/internal/usecase"
)

//TODO HashSlotServer 实现数据分片命令和迁移命令的响应

type HashSlotServer struct {
	Service usecase.IHashSlotService
}

func NewHashSlotServer(serv usecase.IHashSlotService) *HashSlotServer {
	return &HashSlotServer{serv}
}
