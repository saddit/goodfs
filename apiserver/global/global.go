package global

import (
	"github.com/838239178/goodmq"
	"github.com/irfansharif/cfilter"
)

var (
	AmqpConnection *goodmq.AmqpConnection
	// ExistFilter TODO 改造为可持久化的布谷鸟过滤器
	ExistFilter *cfilter.CFilter
)
