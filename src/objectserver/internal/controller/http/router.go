package http

import (
	"common/logs"
	"common/util"
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"objectserver/internal/controller/grpc"
	"objectserver/internal/controller/http/objects"
	"objectserver/internal/controller/http/stat"
	"objectserver/internal/controller/http/temp"
)

type Server struct {
	*http.Server
	grpcServer *grpc.Server
}

func NewHttpServer(port string, grpcServer *grpc.Server) *Server {
	r := gin.New()
	r.Use(gin.LoggerWithWriter(logs.Std().Out), gin.RecoveryWithWriter(logs.Std().Out))
	r.GET("/objects/:name", objects.GetFromCache, objects.Get)
	r.HEAD("/objects/:name", objects.Head)
	r.PUT("/objects/:name", temp.FilterEmptyRequest, objects.Put)
	r.DELETE("/objects/:name", objects.Delete)

	r.POST("/temp/:name", temp.Post)
	r.PATCH("/temp/:name", temp.FilterExpired, temp.FilterEmptyRequest, temp.Patch)
	r.DELETE("/temp/:name", temp.FilterExpired, temp.Delete)
	r.HEAD("/temp/:name", temp.FilterExpired, temp.Head)
	r.GET("/temp/:name", temp.FilterExpired, temp.Get)
	r.PUT("/temp/:name", temp.FilterExpired, temp.Put)

	r.GET("/ping", stat.Ping)
	r.GET("/stat", stat.Info)

	return &Server{&http.Server{
		Addr:    ":" + port,
		Handler: util.H2CHandler(r, grpcServer),
	}, grpcServer}
}

func (h *Server) ListenAndServe() error {
	logs.Std().Infof("http server listen on: %s", h.Addr)
	return h.Server.ListenAndServe()
}

func (s *Server) Shutdown(c context.Context) error {
	var err1, err2 error
	dg := util.NewWaitGroup()
	// shutdown grpc
	dg.Todo()
	go func() {
		defer dg.Done()
		err1 = s.grpcServer.Shutdown(c)
	}()
	// shutdown http
	dg.Todo()
	go func() {
		defer dg.Done()
		err2 = s.Server.Shutdown(c)
	}()
	return errors.Join(err1, err2)
}
