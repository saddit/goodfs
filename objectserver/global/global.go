package global

import (
	"goodfs/lib/util/cache"
	"goodfs/objectserver/config"

	"github.com/838239178/goodmq"
)

var (
	Config         config.Config
	AmqpConnection *goodmq.AmqpConnection
	Cache          *cache.Cache
)
