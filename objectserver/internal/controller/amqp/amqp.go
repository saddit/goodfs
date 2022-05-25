package amqp

import (
	"objectserver/internal/controller/amqp/heartbeat"
	"objectserver/internal/controller/amqp/locate"
)

func Start() {
	go heartbeat.StartHeartbeat()
	go locate.StartLocate()
}
