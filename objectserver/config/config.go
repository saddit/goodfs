package config

const (
	Port         = 8080
	StoragePath  = "E:/file/objects/"
	AmqpAddress  = "amqp://gdfs:gdfs@120.79.59.75:5672/goodfs"
	BeatInterval = 5
)

var (
	LocalAddr string
)
