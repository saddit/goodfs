package auth

import (
	"common/logs"
	"common/response"
	"strings"

	"github.com/gin-gonic/gin"
)

const MiddleKey = "IsAuthenticated"
const MiddleKeyMessage = "AuthenticatedMessage"

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

func AuthenticationMiddleware(cfg *Config, authenticators ...Verification) gin.HandlersChain {
	chain := []gin.HandlerFunc{PreAuthenticate(cfg)}
	for _, v := range authenticators {
		chain = append(chain, AuthenticateWrap(v))
	}
	return append(chain, AfterAuthenticate)
}
