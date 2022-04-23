package global

import (
	"github.com/838239178/cfilter"
	"github.com/838239178/goodmq"
	"net/http"
)

var (
	AmqpConnection *goodmq.AmqpConnection
	// ExistFilter TODO 新节点上线需要能够从其他节点同步此过滤器
	ExistFilter *cfilter.CFilter
	Http        *http.Client
)
