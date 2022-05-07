package locate

import (
	"goodfs/lib/util"
	"goodfs/objectserver/config"
	"goodfs/objectserver/global"
	"goodfs/objectserver/service"
	"log"
	"os"
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
						Body: []byte(config.LocalAddr),
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

func WarmUpLocateCache() {
	files, e := os.ReadDir(global.Config.StoragePath)
	if e != nil {
		panic(e)
	}
	for _, f := range files {
		if !f.IsDir() {
			service.MarkExist(f.Name())
		}
	}
}
