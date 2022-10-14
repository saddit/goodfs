package auth

import (
	"apiserver/config"
	"common/logs"
	"common/response"

	"github.com/gin-gonic/gin"
)

const MiddleKey = "IsAuthenticated"
const MiddleKeyMessage = "AuthenticatedMessage"

func PreAuthenticate(cfg *config.AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// if no enable auth, every request is valid
		c.Set(MiddleKey, !cfg.Enable)
	}
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

func AuthenticationMiddleware(cfg *config.AuthConfig, authenticators []Verification) gin.HandlersChain {
	chain := []gin.HandlerFunc{PreAuthenticate(cfg)}
	for _, v := range authenticators {
		chain = append(chain, AuthenticateWrap(v))
	}
	return append(chain, AfterAuthenticate)
}
