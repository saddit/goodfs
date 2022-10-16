package auth

import (
	"common/logs"
	"common/response"
	"strings"

	"github.com/gin-gonic/gin"
)

const MiddleKey = "IsAuthenticated"
const MiddleErr = "AuthenticatedErr"

func PreAuthenticate(cfg *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// if no enable auth, every request is valid
		if !cfg.Enable {
			c.Set(MiddleKey, true)
			return
		}
		// filter white list
		for _, pref := range cfg.whiteList {
			if strings.HasPrefix(c.FullPath(), pref) {
				c.Set(MiddleKey, true)
				return
			}
		}
		// no within white list
		c.Set(MiddleKey, false)
	}
}

func AuthenticateWrap(validator Verification) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetBool(MiddleKey) {
			return
		}
		if err := validator.Middleware(c); err != nil {
			logs.Std().Infof("authenticate fail: %s", err)
			c.Set(MiddleErr, err)
			return
		}
		c.Set(MiddleKey, true)
	}
}

func AfterAuthenticate(c *gin.Context) {
	if !c.GetBool(MiddleKey) {
		err, _ := c.Get(MiddleErr)
		response.FailErr(err.(error), c).Abort()
	}
}

func AuthenticationMiddleware(cfg *Config, authenticators ...Verification) gin.HandlersChain {
	chain := []gin.HandlerFunc{PreAuthenticate(cfg)}
	for _, v := range authenticators {
		chain = append(chain, AuthenticateWrap(v))
	}
	return append(chain, AfterAuthenticate)
}
