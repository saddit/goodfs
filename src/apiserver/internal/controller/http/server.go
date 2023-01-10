package http

import (
	"apiserver/internal/controller/http/big"
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
	authRoute := eng.Group("/v1", authMid...)

	//rest api
	objects.NewObjectsController(o, m).Register(authRoute)
	big.NewBigObjectsController(o, m).Register(authRoute)
	NewLocateController(o).Register(authRoute)
	NewMetadataController(m).Register(authRoute)
	NewSecurityController().Register(authRoute)

	return &Server{netHttp.Server{Addr: addr, Handler: eng}}
}
