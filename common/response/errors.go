package response

import "fmt"

type IResponseErr interface {
	error
	GetMessage() string
	GetStatus() int
	GetSubMessage() string
}

type ResponseErr struct {
	Status  int
	Message string
}

func (r ResponseErr) Error() string {
	return r.GetMessage()
}

func NewError(code int, msg string) *ResponseErr {
	return &ResponseErr{code, msg}
}

func (r ResponseErr) GetMessage() string {
	return r.Message
}

func (r ResponseErr) GetSubMessage() string {
	return fmt.Sprintf("%T: %s", r, r)
}

func (r ResponseErr) GetStatus() int {
	return r.Status
}
