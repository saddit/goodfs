package heartbeat

import (
	"goodfs/api/config"
	"goodfs/api/service"
	"goodfs/lib/rabbitmq"
	"log"
	"strconv"
	"time"
)

func ListenHearbeat() {
	mq := rabbitmq.New(config.AmqpAddress)
	mq.DeclareQueue("heartbeat.queue")
	mq.CreateBind("apiServers", "heartbeat.queue")
	consumeChan := mq.Consume("")

	go removeExpiredDataServer()

	//no ack to keep connection alive
	for msg := range consumeChan {
		dataServIp, e := strconv.Unquote(string(msg.Body))
		if e != nil {
			log.Printf("Consume heartbeat from data server fail, %v\n", e)
		} else {
			// log.Printf("Receive heartbeat from %v\n", dataServIp)
			service.ReceiveDataServer(dataServIp)
		}
	}
}

//每隔一段时间移除长时间未响应的 data server
func removeExpiredDataServer() {
	for range time.Tick(config.DetectInterval * time.Second) {
		service.CheckServerState()
	}
}
