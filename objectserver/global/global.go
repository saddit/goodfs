package global

import (
	"goodfs/util/cache"

	"github.com/838239178/goodmq"
)

var (
	AmqpConnection *goodmq.AmqpConnection
	Cache          *cache.Cache
)
