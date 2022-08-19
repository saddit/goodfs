package usecase

import (
	"fmt"
	"common/response"
)

var (
	KnownError    = response.NewError(400, "request fail ")
	ErrNotFound = fmt.Errorf("%wnot found", KnownError)
	ErrDBClosed = fmt.Errorf("%wdb closed", KnownError)
	ErrOldData  = fmt.Errorf("%wexpired data", KnownError)
	ErrExists   = fmt.Errorf("%walready exists key", KnownError)
	ErrNilData  = fmt.Errorf("%wnil data", KnownError)
	ErrDecode   = fmt.Errorf("%wdecode fail", KnownError)
	ErrEncode   = fmt.Errorf("%wencode fail", KnownError)
)
