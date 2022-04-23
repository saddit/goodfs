package heartbeat

import (
	"goodfs/apiserver/global"
	"goodfs/apiserver/service"
	"goodfs/lib/util"
	"log"
	"strconv"
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
	consumeChan, ok := consumer.Consume()

	go removeExpiredDataServer()

	//断线重连策略
	for range util.ImmediateTick(5 * time.Second) {
		if ok {
			log.Println("Heartbeat connect success")
			for msg := range consumeChan {
				dataServIp, e := strconv.Unquote(string(msg.Body))
				if e != nil {
					log.Printf("Consume heartbeat from data server fail, %v\n", e)
				} else {
					//log.Printf("Receive heartbeat from %v\n", dataServIp)
					service.ReceiveDataServer(dataServIp)
				}
			}
			ok = false
		} else {
			log.Println("Heartbeat connection closed! Recovering...")
			//重试直到成功
			consumeChan, ok = consumer.Consume()
		}
	}
}

//每隔一段时间移除长时间未响应的 data server
func removeExpiredDataServer() {
	for range time.Tick(global.Config.DetectInterval) {
		service.CheckServerState()
	}
}
