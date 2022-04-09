package service

import "errors"

var (
	ErrServiceUnavailable = errors.New("DataServer unavailable")
	ErrInternalServer     = errors.New("Internal server error")
	ErrBadRequest         = errors.New("Bad Request")
)
