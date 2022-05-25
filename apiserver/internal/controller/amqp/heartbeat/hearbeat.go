package heartbeat

import (
	"apiserver/internal/usecase/pool"
	"apiserver/internal/usecase/service"
	"time"

	"github.com/streadway/amqp"
)

func ListenHeartbeat() {
	consumer, err := pool.Amqp.NewConsumer()
	if err != nil {
		panic(err)
	}
	defer consumer.Close()
	consumer.DeleteUnused = true
	consumer.Exchange = "apiServers"

	go removeExpiredDataServer()

	consumer.ConsumeAuto(func(msg amqp.Delivery) {
		service.ReceiveDataServer(string(msg.Body))
	}, 5*time.Second)
}

//每隔一段时间移除长时间未响应的 data server
func removeExpiredDataServer() {
	for range time.Tick(pool.Config.Discovery.DetectInterval) {
		service.CheckServerState()
	}
}
