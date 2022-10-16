package response

type RaftFsmResp struct {
	ResponseErr
	Data any
}

func NewRaftFsmResp(err error) *RaftFsmResp {
	switch err := err.(type) {
	case *ResponseErr:
		return &RaftFsmResp{*err, nil}
	case ResponseErr:
		return &RaftFsmResp{err, nil}
	case nil:
		return &RaftFsmResp{ResponseErr{Status: 200}, nil}
	default:
		return &RaftFsmResp{ResponseErr{Status: 500, Message: err.Error()}, nil}
	}
}

func (r *RaftFsmResp) Ok() bool {
	return r.Status/100 == 2
}
