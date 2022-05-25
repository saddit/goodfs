package objects

import (
	. "apiserver/internal/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var (
	ObjectService IObjectService
	MetaService   IMetaService
)

func AbortInternalError(c *gin.Context, err error) {
	logrus.Errorln(c.AbortWithError(http.StatusInternalServerError, err))
}

func AbortServiceUnavailableError(c *gin.Context, err error) {
	logrus.Errorln(c.AbortWithError(http.StatusServiceUnavailable, err))
}
