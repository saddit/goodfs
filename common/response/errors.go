package response

import "fmt"

type ResponseErr struct {
	Status  int
	Message string
}

func (r ResponseErr) Error() string {
	return fmt.Sprint(r.Status, ":", r.Message)
}

func NewError(code int, msg string) *ResponseErr{
	return &ResponseErr{code, msg}
}
