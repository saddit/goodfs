package usecase

import (
	"errors"
	"fmt"
)

var (
	ErrKnown    = errors.New("")
	ErrNotFound = fmt.Errorf("%wnot found", ErrKnown)
	ErrOldData  = fmt.Errorf("%wexpired data", ErrKnown)
	ErrExists   = fmt.Errorf("%walready exists key", ErrKnown)
	ErrNilData  = fmt.Errorf("%wnil data", ErrKnown)
	ErrDecode   = fmt.Errorf("%wdecode fail", ErrKnown)
	ErrEncode   = fmt.Errorf("%wencode fail", ErrKnown)
)
