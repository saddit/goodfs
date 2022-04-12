package heartbeat

import (
	"goodfs/api/config"
	"goodfs/api/service"
	"log"
	"strconv"
	"time"

	"github.com/838239178/goodmq"
)

func ListenHeartbeat() {
	mq := goodmq.NewAmqpConnection(config.AmqpAddress)
	consumer, err := mq.NewConsumer()
	if err != nil {
		panic(err)
	}
	defer consumer.Close()
	consumer.QueName = "heartbeat.queue"
	consumer.Exchange = "apiServers"
	consumeChan, ok := consumer.Consume()
	// mq := rabbitmq.New(config.AmqpAddress)
	// mq.DeclareQueue("heartbeat.queue")
	// mq.CreateBind("apiServers", "heartbeat.queue")
	// consumeChan := mq.Consume("")

	go removeExpiredDataServer()

	//断线重连策略
	for range time.Tick(5 * time.Second) {
		if ok {
			log.Println("Hearbeat connect success")
			for msg := range consumeChan {
				dataServIp, e := strconv.Unquote(string(msg.Body))
				if e != nil {
					log.Printf("Consume heartbeat from data server fail, %v\n", e)
				} else {
					// log.Printf("Receive heartbeat from %v\n", dataServIp)
					service.ReceiveDataServer(dataServIp)
				}
			}
			ok = false
		} else {
			log.Println("Hearbeat connection closed! Recovering...")
			//重试直到成功
			consumeChan, ok = consumer.Consume()
		}
	}
}

//每隔一段时间移除长时间未响应的 data server
func removeExpiredDataServer() {
	for range time.Tick(config.DetectInterval * time.Second) {
		service.CheckServerState()
	}
}
