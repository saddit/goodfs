package amqp

import (
	"apiserver/config"
	"apiserver/internal/controller/amqp/heartbeat"

	"github.com/838239178/goodmq"
)

func Start(cfg config.DiscoveryConfig, conn *goodmq.AmqpConnection) {
	go heartbeat.ListenHeartbeat(cfg, conn)
}
