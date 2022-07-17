package amqp

import (
	"objectserver/internal/controller/amqp/locate"
)

func Start() {
	go locate.StartLocate()
}
