package response

type RaftFsmResp struct {
	*Err
	Data any
}

func NewRaftFsmResp(err error) *RaftFsmResp {
	switch err := err.(type) {
	case *Err:
		return &RaftFsmResp{err, nil}
	case Err:
		return &RaftFsmResp{&err, nil}
	case nil:
		return &RaftFsmResp{&Err{Status: 200}, nil}
	default:
		return &RaftFsmResp{&Err{Status: 500, Message: err.Error()}, nil}
	}
}

func (r *RaftFsmResp) Ok() bool {
	return r.Status/100 == 2
}

func (r *RaftFsmResp) Unwrap() error {
	return r.Err
}
