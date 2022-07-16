package heartbeat

import (
	"apiserver/config"
	"apiserver/internal/usecase/service"
	"time"

	"github.com/838239178/goodmq"
	"github.com/streadway/amqp"
)

func ListenHeartbeat(cfg config.DiscoveryConfig, conn *goodmq.AmqpConnection) {
	consumer, err := conn.NewConsumer()
	if err != nil {
		panic(err)
	}
	defer consumer.Close()
	consumer.DeleteUnused = true
	consumer.Exchange = "apiServers"

	go removeExpiredDataServer(cfg.DetectInterval)

	consumer.ConsumeAuto(func(msg amqp.Delivery) {
		service.ReceiveDataServer(string(msg.Body))
	}, 5*time.Second)
}

//每隔一段时间移除长时间未响应的 data server
func removeExpiredDataServer(detect time.Duration) {
	for range time.Tick(detect) {
		service.CheckServerState()
	}
}
