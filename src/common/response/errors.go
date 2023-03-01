package response

import "fmt"

type IErr interface {
	error
	GetMessage() string
	GetStatus() int
	GetSubMessage() string
}

type Err struct {
	Status  int
	Message string
}

func (r Err) Error() string {
	return r.GetMessage()
}

func NewError(code int, msg string) *Err {
	return &Err{code, msg}
}

func (r Err) GetMessage() string {
	return r.Message
}

func (r Err) GetSubMessage() string {
	return fmt.Sprintf("%T: %s", r, r)
}

func (r Err) GetStatus() int {
	return r.Status
}

func CheckErrStatus(status int, err error) bool {
	if err == nil {
		return false
	}
	respErr, ok := err.(IErr)
	if !ok {
		return false
	}
	return respErr.GetStatus() == status
}
