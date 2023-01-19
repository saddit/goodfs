package http

import (
	"adminserver/internal/usecase/logic"
	"adminserver/internal/usecase/webapi"
	"common/logs"
	"common/response"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const (
	authTokenKey = "AuthToken"
)

func SaveToken(c *gin.Context) {
	sess := sessions.Default(c)
	originToken := sess.Get(authTokenKey)
	authToken := c.GetHeader("Authorization")
	if originToken != nil && originToken.(string) == authToken {
		return
	}
	sess.Set(authTokenKey, authToken)
	if err := sess.Save(); err != nil {
		logs.Std().Error(err)
	}
}

func GetAuthToken(c *gin.Context) string {
	tk := sessions.Default(c).Get(authTokenKey)
	if tk == nil {
		return c.GetHeader("Authorization")
	}
	return tk.(string)
}

func RequireToken(c *gin.Context) {
	if GetAuthToken(c) == "" {
		response.Exec(c).Abort().Fail(403, "invalid credential")
		return
	}
	c.Next()
}

func ClearCredential(router gin.IRouter) {
	router.POST("/logout", func(c *gin.Context) {
		sessions.Default(c).Clear()
		response.Ok(c)
	})
}

func CheckCredential(router gin.IRouter) {
	router.POST("/login", func(c *gin.Context) {
		token := GetAuthToken(c)
		err := webapi.CheckToken(logic.SelectApiServer(), token)
		if err != nil {
			response.FailErr(err, c)
			return
		}
		response.Ok(c)
	})
}
