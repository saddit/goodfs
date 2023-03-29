package http

import (
	"apiserver/internal/usecase/pool"
	"common/response"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
	"net/http"
)

func Ping(c *gin.Context) {
	c.Status(http.StatusOK)
}

func Config(c *gin.Context) {
	conf := *pool.Config
	conf.Etcd.Username = "*****"
	conf.Etcd.Password = "*****"
	conf.Auth.Password.Password = "*****"
	conf.Auth.Password.Username = "*****"
	out, err := yaml.Marshal(&conf)
	if err != nil {
		response.FailErr(err, c)
		return
	}
	_, _ = c.Writer.Write(out)
	response.Ok(c)
}
