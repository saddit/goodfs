package raftimpl

import (
	"common/response"
	"common/util"
	"metaserver/internal/entity"
	"metaserver/internal/usecase"
	"time"
)

type raftApplier struct {
	wrapper *RaftWrapper
}

func RaftApplier(wrapper *RaftWrapper) usecase.RaftApply {
	return &raftApplier{wrapper: wrapper}
}

func (r *raftApplier) ApplyRaft(data *entity.RaftData) (bool, *response.RaftFsmResp) {
	if rf, ok := r.wrapper.GetRaftIfLeader(); ok {
		bt, err := util.EncodeMsgp(data)
		if err != nil {
			return true, response.NewRaftFsmResp(err)
		}
		feat := rf.Apply(bt, 5*time.Second)
		if err := feat.Error(); err != nil {
			return true, response.NewRaftFsmResp(err)
		}
		if resp := feat.Response(); resp != nil {
			return true, resp.(*response.RaftFsmResp)
		}
		return true, nil
	}
	return false, nil
}
