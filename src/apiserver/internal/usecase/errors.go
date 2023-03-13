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
	ErrMetadataExists     = response.NewError(http.StatusInternalServerError, "metadata exist")
	ErrInvalidFile        = response.NewError(http.StatusBadRequest, "invalid file")
	ErrOverRead           = errors.New("read to much data")
	ErrStreamClosed       = errors.New("stream closed")
)
