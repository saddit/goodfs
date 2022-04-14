package config

import "goodfs/util/datasize"

const (
	Port         = 8080
	StoragePath  = "E:/file/objects/"
	AmqpAddress  = "amqp://gdfs:gdfs@120.79.59.75:5672/goodfs"
	BeatInterval = 5
	CacheSize    = 64 * datasize.MB
)

var (
	LocalAddr string
)
