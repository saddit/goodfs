package auth

import "github.com/gin-gonic/gin"

type (
	Credential interface {
		GetUsername() string
		GetPassword() string
		GetExtra() map[string][]string
	}

	IAuthenticator interface {
		ProviderVerification() Verification
	}

	Verification interface {
		Verify(Credential) error
		Middleware(*gin.Context) error
	}
)
