package locate

import (
	"goodfs/lib/rabbitmq"
	"goodfs/objects/config"
	"goodfs/objects/service"
	"log"
	"strconv"
)

/*
	开始监听对象寻址消息队列
*/
func StartLocate() {
	mq := rabbitmq.New(config.AmqpAddress)
	defer mq.Close()

	mq.DeclareQueue("data.locate.queue")
	mq.Bind("dataServers")

	consumeChan := mq.Consume("")

	//no ack to keep connection alive
	for msg := range consumeChan {
		object, e := strconv.Unquote(string(msg.Body))
		if e != nil {
			log.Printf("Locate consume fail, %v\n", e)
		} else if service.Exist(object) {
			mq.Send(msg.ReplyTo, config.LocalAddr)
		}
	}
}
