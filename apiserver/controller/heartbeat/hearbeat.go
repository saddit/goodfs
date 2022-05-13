package heartbeat

import (
	"github.com/streadway/amqp"
	"goodfs/apiserver/global"
	"goodfs/apiserver/service/dataserv"
	"time"
)

func ListenHeartbeat() {
	consumer, err := global.AmqpConnection.NewConsumer()
	if err != nil {
		panic(err)
	}
	defer consumer.Close()
	consumer.DeleteUnused = true
	consumer.Exchange = "apiServers"

	go removeExpiredDataServer()

	consumer.ConsumeAuto(func(msg amqp.Delivery) {
		dataserv.ReceiveDataServer(string(msg.Body))
	}, 5*time.Second)
}

//每隔一段时间移除长时间未响应的 data server
func removeExpiredDataServer() {
	for range time.Tick(global.Config.DetectInterval) {
		dataserv.CheckServerState()
	}
}
