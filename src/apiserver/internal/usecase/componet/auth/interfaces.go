package auth

import "github.com/gin-gonic/gin"

type (
	Credential interface {
		GetUsername() string
		GetPassword() string
		GetExtra() map[string][]string
	}

	Verification interface {
		Verify(Credential) error
		Middleware(*gin.Context) (bool, error)
	}
)
