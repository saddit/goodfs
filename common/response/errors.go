package response

type ResponseErr struct {
	Status  int
	Message string
}

func (r ResponseErr) Error() string {
	return r.Message
}

func NewError(code int, msg string) *ResponseErr{
	return &ResponseErr{code, msg}
}
