package heartbeat

import (
	"encoding/json"
	"goodfs/objects/config"
	"goodfs/objects/global"
	"time"

	"github.com/streadway/amqp"
)

func StartHeartbeat() {
	sender, err := global.AmqpConnection.NewProvider()
	if err != nil {
		panic(err)
	}
	defer sender.Close()
	sender.Exchange = "apiServers"
	locate, _ := json.Marshal(config.LocalAddr)

	for range time.Tick(config.BeatInterval * time.Second) {
		// log.Println("Send Heartbeat")
		sender.Publish(amqp.Publishing{
			Body: locate,
		})
	}
}
