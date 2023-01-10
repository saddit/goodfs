package http

import (
	"common/logs"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const (
	authTokenKey = "AuthToken"
)

func SaveToken(c *gin.Context) {
	sess := sessions.Default(c)
	authToken := c.GetHeader("Authorization")
	sess.Set(authTokenKey, authToken)
	if err := sess.Save(); err != nil {
		logs.Std().Error(err)
	}
}

func GetAuthToken(c *gin.Context) string {
	tk := sessions.Default(c).Get(authTokenKey)
	if tk == nil {
		return ""
	}
	return tk.(string)
}
