package controller

import (
	http2 "adminserver/internal/controller/http"
	"common/logs"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"strings"
)

type HttpServer struct {
	http.Server
}

func NewHttpServer(addr string, webFs static.ServeFileSystem) *HttpServer {
	eng := gin.Default()

	randSec := uuid.New()
	eng.Use(static.Serve("/", webFs))
	eng.Use(sessions.Sessions("dfs-admin", cookie.NewStore(randSec[:])))
	eng.Use(http2.SaveToken)
	eng.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		// AllowOrigins: []string{"http://localhost", "http://localhost:5173"},
		AllowMethods:  []string{"PUT", "PATCH", "POST", "GET", "OPTION"},
		AllowHeaders:  []string{"Authorization", "Content-Type", "Accept", "Refer"},
		ExposeHeaders: []string{"X-Total-Count"},
	}))

	route := eng.Group("/api")
	http2.NewMetadataController().Register(route)
	http2.NewServerStateController().Register(route)
	http2.NewObjectsController().Register(route)

	// redirect to index of console if no route
	eng.NoRoute(func(c *gin.Context) {
		url := c.Request.URL.Path
		if strings.HasPrefix(url, "/api") {
			return
		}
		if url == "/favicon.ico" {
			return
		}
		c.Redirect(http.StatusPermanentRedirect, "/")
	})

	return &HttpServer{Server: http.Server{Handler: eng, Addr: addr}}
}

func (s *HttpServer) ListenAndServe() error {
	logs.New("http-server").Infof("server listening on %s", s.Addr)
	return s.Server.ListenAndServe()
}
