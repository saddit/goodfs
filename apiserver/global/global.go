package global

import (
	"github.com/838239178/goodmq"
	"github.com/irfansharif/cfilter"
)

var (
	AmqpConnection *goodmq.AmqpConnection
	ExistFilter    *cfilter.CFilter
)
