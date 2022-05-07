package service

import "errors"

type KnownErr error

var (
	ErrServiceUnavailable KnownErr = errors.New("dataServer unavailable")
	ErrInternalServer     KnownErr = errors.New("internal server error")
	ErrBadRequest         KnownErr = errors.New("bad Request")
	ErrInvalidFile        KnownErr = errors.New("invalid file")
)
