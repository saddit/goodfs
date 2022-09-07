package response

type IResponseErr interface {
	error
	GetMessage() string
	GetStatus() int
}

type ResponseErr struct {
	Status  int
	Message string
}

func (r ResponseErr) Error() string {
	return r.Message
}

func NewError(code int, msg string) *ResponseErr {
	return &ResponseErr{code, msg}
}

func (r ResponseErr) GetMessage() string {
	return r.Message
}

func (r ResponseErr) GetStatus() int {
	return r.Status
}
