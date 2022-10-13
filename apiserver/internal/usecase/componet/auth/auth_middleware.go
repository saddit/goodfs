package auth

import (
	"apiserver/internal/usecase/pool"
	"common/logs"
	"common/response"
	"common/util"
	"github.com/gin-gonic/gin"
)

const MiddleKey = "IsAuthenticated"
const MiddleKeyMessage = "AuthenticatedMessage"

func PreAuthenticate(c *gin.Context) {
	isAuth := util.IfElse(pool.Config.Auth.Enable, false, true)
	c.Set("AuthKey", isAuth)
}

func AuthenticateWrap(validator Verification) gin.HandlerFunc {
	return func(c *gin.Context) {
		if res, _ := c.Get(MiddleKey); res == true {
			return
		}
		if err := validator.Middleware(c); err != nil {
			logs.Std().Infof("authenticate fail: %s", err)
			c.Set(MiddleKeyMessage, err.Error())
			return
		}
		c.Set(MiddleKey, true)
	}
}

func AfterAuthenticate(c *gin.Context) {
	res, _ := c.Get(MiddleKey)
	msg, _ := c.Get(MiddleKeyMessage)
	if res == false {
		response.FailErr(response.NewError(403, msg.(string)), c).Abort()
	}
}

func AuthenticationMiddleware() gin.HandlersChain {
	chain := []gin.HandlerFunc{PreAuthenticate}
	for _, v := range pool.Authenticators {
		chain = append(chain, AuthenticateWrap(v))
	}
	return append(chain, AfterAuthenticate)
}
