package locate

import (
	"encoding/json"
	"goodfs/objectserver/config"
	"goodfs/objectserver/global"
	"goodfs/objectserver/service"
	"goodfs/util"
	"log"
	"strconv"
	"time"

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

	consumeChan, ok := conm.Consume()

	for range util.ImmediateTick(5 * time.Second) {
		if ok {
			log.Println("Start locate server")
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
			ok = false
		} else {
			log.Println("Oops! Recovering locate server")
			consumeChan, ok = conm.Consume()
		}
	}
}
