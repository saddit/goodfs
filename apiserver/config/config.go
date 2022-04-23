package config

const (
	Port           = 8082
	AmqpAddress    = "amqp://gdfs:gdfs@120.79.59.75:5672/goodfs"
	MongoAddress   = "mongodb://150.158.82.154:27017#study#SCRAM-SHA-256#root#xianka"
	LogDir         = "e:/file/logs"
	DetectInterval = 5
	// SuspendTimeout NumPerDetect   = 5
	SuspendTimeout = 5
	DeadTimeout    = 10
	SelectStrategy = "random"
	MachineCode    = "1"
)
