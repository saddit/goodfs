package http

import (
	"apiserver/config"
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
	tls *config.TLSConfig
}

func NewHttpServer(addr string, o IObjectService, m IMetaService, b repo.IBucketRepo) *Server {
	authMid := auth.AuthenticationMiddleware(&pool.Config.Auth,
		auth.NewCallbackValidator(&pool.Config.Auth.Callback),
		auth.NewPasswordValidator(pool.Etcd, &pool.Config.Auth.Password),
	)

	eng := gin.New()
	eng.Use(gin.LoggerWithWriter(logs.Std().Out), gin.RecoveryWithWriter(logs.Std().Out))
	eng.UseRawPath = true
	eng.UnescapePathValues = false

	eng.GET("/ping", func(c *gin.Context) { _, _ = c.Writer.Write([]byte("pong")) })

	authRoute := eng.Group("/v1", authMid...)
	{
		NewObjectsController(o, m).Register(authRoute)
		NewBigObjectsController(o, m, b).Register(authRoute)
		NewLocateController(o).Register(authRoute)
		NewMetadataController(m).Register(authRoute)
		NewSecurityController().Register(authRoute)
		NewBucketController(b).Register(authRoute)
	}

	return &Server{netHttp.Server{Addr: addr, Handler: eng}, &pool.Config.TLS}
}

func (s *Server) ListenAndServe() error {
	logs.Std().Infof("http server listen on: %s", s.Server.Addr)
	if s.tls.Enabled {
		logs.Std().Infof("tls enabled, use cert: %s, key: %s", s.tls.ServerCertFile, s.tls.ServerKeyFile)
		return s.Server.ListenAndServeTLS(s.tls.ServerCertFile, s.tls.ServerKeyFile)
	}
	return s.Server.ListenAndServe()
}
