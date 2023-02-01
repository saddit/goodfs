package usecase

import (
	"common/response"
	"errors"
	"net/http"
)

var (
	ErrServiceUnavailable = response.NewError(http.StatusServiceUnavailable, "dataServer unavailable")
	ErrInternalServer     = response.NewError(http.StatusInternalServerError, "internal server error")
	ErrNotFound           = response.NewError(http.StatusNotFound, "resource not found")
	ErrBadRequest         = response.NewError(http.StatusBadRequest, "bad Request")
	ErrInvalidFile        = response.NewError(http.StatusBadRequest, "invalid file")
	ErrNeedUpdateMeta     = response.NewError(http.StatusBadRequest, "metadata has changed unavailable server's location")
	ErrOverRead           = errors.New("read to much data")
)
