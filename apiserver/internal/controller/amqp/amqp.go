package amqp

import "apiserver/internal/controller/amqp/heartbeat"

func Start() {
	go heartbeat.ListenHeartbeat()
}
