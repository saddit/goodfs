package config

import (
	"goodfs/util/datasize"
	"time"
)

const (
	Port               = 8080
	StoragePath        = "E:/file/objects/"
	TempPath           = "E:/file/objects/temp/"
	AmqpAddress        = "amqp://gdfs:gdfs@120.79.59.75:5672/goodfs"
	BeatInterval       = 5
	CacheMaxSizeMB       = 256
	CacheTTL           = 1 * time.Hour
	CacheCleanInterval = CacheTTL / 10
	CacheItemMaxSize   = 32 * datasize.MB
)

var (
	LocalAddr string
)
