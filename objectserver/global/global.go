package global

import (
	"github.com/838239178/goodmq"
	"github.com/VictoriaMetrics/fastcache"
)

var (
	AmqpConnection *goodmq.AmqpConnection
	Cache          *fastcache.Cache
)
