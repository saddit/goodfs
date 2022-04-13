package locate

import (
	"encoding/json"
	"goodfs/objects/config"
	"goodfs/objects/global"
	"goodfs/objects/service"
	"log"
	"strconv"

	"github.com/streadway/amqp"
)

/*
	开始监听对象寻址消息队列
*/
func StartLocate() {
	conm, e := global.AmqpConnection.NewConsumer()
	if e != nil {
		panic(e)
	}
	defer conm.Close()
	conm.Exchange = "dataServers"
	conm.DeleteUnused = true

	prov, e := global.AmqpConnection.NewProvider()
	if e != nil {
		panic(e)
	}
	defer prov.Close()

	locate, e := json.Marshal(config.LocalAddr)
	if e != nil {
		panic(e)
	}

	if consumeChan, ok := conm.Consume(); ok {
		for msg := range consumeChan {
			object, e := strconv.Unquote(string(msg.Body))
			if e != nil {
				log.Printf("Locate consume fail, %v\n", e)
			} else if service.Exist(object) {
				prov.RouteKey = msg.ReplyTo
				prov.Publish(amqp.Publishing{
					Body: locate,
				})
			}
		}
	} else {
		panic("Consume Locate Error")
	}
}
