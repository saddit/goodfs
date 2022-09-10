package usecase

import (
	"common/response"
	"net/http"
)

var (
	ErrServiceUnavailable = response.NewError(http.StatusServiceUnavailable, "dataServer unavailable")
	ErrInternalServer     = response.NewError(http.StatusInternalServerError, "internal server error")
	ErrBadRequest         = response.NewError(http.StatusBadRequest, "bad Request")
	ErrInvalidFile        = response.NewError(http.StatusBadRequest, "invalid file")
	ErrNeedUpdateMeta     = response.NewError(http.StatusBadRequest, "metadata has changed unavailable server's location")
)
