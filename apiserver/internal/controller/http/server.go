package http

import (
	"apiserver/internal/controller/http/big"
	"apiserver/internal/controller/http/locate"
	"apiserver/internal/controller/http/objects"
	. "apiserver/internal/usecase"
	"apiserver/internal/usecase/componet/auth"
	"apiserver/internal/usecase/pool"
	netHttp "net/http"

	"github.com/gin-gonic/gin"
)

type Server struct {
	netHttp.Server
}

func NewHttpServer(addr string, o IObjectService, m IMetaService) *Server {
	authMid := auth.AuthenticationMiddleware(&pool.Config.Auth,
		auth.NewCallbackValidator(pool.Http, &pool.Config.Auth.Callback),
		auth.NewPasswordValidator(pool.Etcd, &pool.Config.Auth.Password),
	)

	eng := gin.Default()
	r := eng.Group("/v1").Use(authMid...)

	objects.NewObjectsControoler(o, m).Register(r)
	big.NewBigObjectsController(o, m).Register(r)
	locate.NewLocateController(o).Register(r)
	
	return &Server{netHttp.Server{Addr: addr, Handler: eng}}
}