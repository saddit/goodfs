package http

import (
	"common/graceful"
	. "metaserver/internal/usecase"
	hhttp "net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/raft"
	"google.golang.org/grpc"
)

type HttpServer struct {
	hhttp.Handler
	addr string
}

func NewHttpServer(addr string, grpcServer *grpc.Server, service IMetadataService, rf *raft.Raft) *HttpServer {
	engine := gin.Default()
	if grpcServer != nil {
		//grpc router
		engine.Use(func(ctx *gin.Context) {
			if ctx.Request.ProtoMajor == 2 &&
				strings.HasPrefix(ctx.GetHeader("Content-Type"), "application/grpc") {
				// 按grpc方式来请求
				grpcServer.ServeHTTP(ctx.Writer, ctx.Request)
				// 不要再往下请求了,防止继续链式调用拦截器
				ctx.Abort()
				return
			}
			ctx.Next()
		})
	}
	//Http router
	mc := NewMetadataController(rf, service)
	engine.PUT("/metadata/{name}", mc.Put)
	engine.POST("/metadata/{name}", mc.Post)
	engine.GET("/metadata/{name}", mc.Get)
	engine.DELETE("/metadata/{name}", mc.Delete)

	vc := NewVersionController(rf, service)
	engine.PUT("/metadata_version/{name}", vc.Put)
	engine.POST("/metadata_version/{name}", vc.Post)
	engine.GET("/metadata_version/{name}", vc.Get)
	engine.DELETE("/metadata_version/{name}", vc.Delete)

	return &HttpServer{engine, addr}
}

func (h *HttpServer) ListenAndServe() {
	graceful.ListenAndServe(h.addr, h)
}
