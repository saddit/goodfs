package locate

import (
	"encoding/json"
	"fmt"
	"goodfs/lib/util"
	"goodfs/objectserver/config"
	"goodfs/objectserver/global"
	"goodfs/objectserver/model"
	"goodfs/objectserver/service"
	"log"
	"os"
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

func SyncExistingFilter() {
	log.Println("Start syncing existing file name...")
	defer log.Println("Finish Syncing existing file name")

	provider, err := global.AmqpConnection.NewProvider()
	if err != nil {
		panic(err)
	}
	defer provider.Close()
	provider.Exchange = "existSync"

	dir, err := os.ReadDir(global.Config.StoragePath)
	if err != nil {
		panic(err)
	}

	for _, entry := range dir {
		if entry.IsDir() {
			continue
		}
		if !provider.Publish(amqp.Publishing{
			Body: []byte(entry.Name()),
			Type: model.SyncInsert,
		}) {
			panic(fmt.Errorf("%v sync fail", entry.Name()))
		}
	}
}
