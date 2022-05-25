package locate

import (
	"common/util"
	"log"
	"objectserver/internal/usecase/pool"
	"objectserver/internal/usecase/service"
	"time"

	"github.com/streadway/amqp"
)

/*
	开始监听对象寻址消息队列
*/
func StartLocate() {
	conm, e := pool.Amqp.NewConsumer()
	if e != nil {
		panic(e)
	}
	defer conm.Close()
	conm.Exchange = "dataServers"
	conm.DeleteUnused = true

	prov, e := pool.Amqp.NewProvider()
	if e != nil {
		panic(e)
	}
	defer prov.Close()

	consumeChan, ok := conm.Consume()

	for range util.ImmediateTick(5 * time.Second) {
		if ok {
			log.Println("Start locate server")
			for msg := range consumeChan {
				hash := string(msg.Body)
				if service.Exist(hash) {
					prov.RouteKey = msg.ReplyTo
					prov.Publish(amqp.Publishing{
						Type: hash,
						Body: []byte(pool.Config.LocalAddr()),
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
