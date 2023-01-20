package http

import (
	"apiserver/internal/controller/http/big"
	"apiserver/internal/controller/http/objects"
	. "apiserver/internal/usecase"
	"apiserver/internal/usecase/componet/auth"
	"apiserver/internal/usecase/pool"
	"apiserver/internal/usecase/repo"
	"common/logs"
	"github.com/gin-gonic/gin"
	netHttp "net/http"
)

type Server struct {
	netHttp.Server
}

func NewHttpServer(addr string, o IObjectService, m IMetaService, b repo.IBucketRepo) *Server {
	authMid := auth.AuthenticationMiddleware(&pool.Config.Auth,
		auth.NewCallbackValidator(pool.Http, &pool.Config.Auth.Callback),
		auth.NewPasswordValidator(pool.Etcd, &pool.Config.Auth.Password),
	)

	eng := gin.New()
	eng.Use(gin.LoggerWithWriter(logs.Std().Out), gin.RecoveryWithWriter(logs.Std().Out))
	eng.UseRawPath = true
	eng.UnescapePathValues = false
	authRoute := eng.Group("/v1", authMid...)

	//rest api
	objects.NewObjectsController(o, m).Register(authRoute)
	big.NewBigObjectsController(o, m, b).Register(authRoute)
	NewLocateController(o).Register(authRoute)
	NewMetadataController(m).Register(authRoute)
	NewSecurityController().Register(authRoute)
	NewBucketController(b).Register(authRoute)

	return &Server{netHttp.Server{Addr: addr, Handler: eng}}
}
