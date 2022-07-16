package usecase

import "errors"

var (
	ErrNotFound = errors.New("not found")
	ErrExists   = errors.New("already exists key")
	ErrNilData  = errors.New("nil data")
	ErrDecode   = errors.New("decode fail")
	ErrEncode   = errors.New("encode fail")
)
