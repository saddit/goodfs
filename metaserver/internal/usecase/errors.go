package usecase

import (
	"common/response"
)

var (
	ErrNotFound = response.NewError(404, "not found")
	ErrDBClosed = response.NewError(502, "database closed")
	ErrOldData  = response.NewError(400, "data expired")
	ErrExists   = response.NewError(400, "data exists")
	ErrNilData  = response.NewError(400, "null value")
	ErrDecode   = response.NewError(500, "data decode fail")
	ErrEncode   = response.NewError(500, "data encode fail")
)
