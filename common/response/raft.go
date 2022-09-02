package response

type RaftFsmResp struct {
	err error
	Data any
}

func NewRaftFsmResp(err error) *RaftFsmResp {
	return &RaftFsmResp{err, nil}
}

func (r RaftFsmResp) Error() string {
	return r.err.Error()
}

func (r *RaftFsmResp) Ok() bool {
	return r.err == nil
}