package heartbeat

import (
	"goodfs/lib/rabbitmq"
	"goodfs/objects/config"
	"log"
	"time"
)

func StartHeartbeat() {
	conn := rabbitmq.New(config.AmqpAddress)
	defer conn.Close()

	for range time.Tick(config.BeatInterval * time.Second) {
		e := conn.Publish("apiServers", config.LocalAddr)
		if e != nil {
			log.Printf("Publish to apiServers exchange error, %v\n", e)
		}
		// log.Println("Send heartbeat!")
	}
}
