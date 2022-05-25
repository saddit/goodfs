package heartbeat

import (
	"common/util"
	"log"
	"objectserver/internal/usecase/pool"

	"github.com/streadway/amqp"
)

func StartHeartbeat() {
	sender, err := pool.Amqp.NewProvider()
	if err != nil {
		panic(err)
	}
	defer sender.Close()
	sender.Exchange = "apiServers"
	log.Println("Start heartbeat..")

	for range util.ImmediateTick(pool.Config.BeatInterval) {
		// log.Println("Send Heartbeat")
		sender.Publish(amqp.Publishing{
			Body: []byte(pool.Config.LocalAddr()),
		})
	}
}
