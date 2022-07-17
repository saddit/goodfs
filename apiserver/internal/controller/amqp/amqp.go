package amqp

import (
	"apiserver/config"
	"apiserver/internal/controller/amqp/heartbeat"

	"github.com/838239178/goodmq"
)

//Start Deprecated 弃用
func Start(cfg config.DiscoveryConfig, conn *goodmq.AmqpConnection) {
	go heartbeat.ListenHeartbeat(cfg, conn)
}
