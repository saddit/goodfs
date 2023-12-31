package raftimpl

import (
	"common/util"
	"fmt"
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

func (r *raftApplier) ApplyRaft(data *entity.RaftData) (bool, any, error) {
	if rf, ok := r.wrapper.GetRaftIfLeader(); ok {
		bt, err := util.EncodeMsgp(data)
		if err != nil {
			return true, nil, err
		}
		feat := rf.Apply(bt, 5*time.Second)
		if err = feat.Error(); err != nil {
			return true, nil, err
		}
		resp := feat.Response()
		if resp == nil {
			return true, nil, fmt.Errorf("unknown response from fsm: %v", resp)
		}
		if rp, ok := resp.(*FSMResponse); ok {
			return true, rp.Data, rp.ToError()
		}
		if rp, ok := resp.(error); ok {
			return true, nil, rp
		}		
	}
	return false, nil, nil
}
