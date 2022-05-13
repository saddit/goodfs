package heartbeat

import (
	"github.com/streadway/amqp"
	"goodfs/lib/util"
	"goodfs/objectserver/config"
	"goodfs/objectserver/global"
	"log"
)

func StartHeartbeat() {
	sender, err := global.AmqpConnection.NewProvider()
	if err != nil {
		panic(err)
	}
	defer sender.Close()
	sender.Exchange = "apiServers"
	log.Println("Start heartbeat..")

	for range util.ImmediateTick(global.Config.BeatInterval) {
		// log.Println("Send Heartbeat")
		sender.Publish(amqp.Publishing{
			Body: []byte(config.LocalAddr),
		})
	}
}
