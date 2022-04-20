package locate

import (
	"github.com/streadway/amqp"
	"goodfs/apiserver/config"
	"goodfs/apiserver/global"
	"goodfs/apiserver/model"
	"goodfs/apiserver/service"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func Get(ctx *gin.Context) {
	name := ctx.Param("name")
	info, exist := service.LocateFile(name)
	if !exist {
		ctx.Status(http.StatusNotFound)
	} else {
		ctx.JSON(http.StatusOK, info)
	}
}

func SyncExistFilter() {
	consumer, err := global.AmqpConnection.NewConsumer()
	if err != nil {
		panic(err)
	}
	defer consumer.Close()
	//消费失败的情况需要持久化消息，保证能够恢复到一致性状态
	consumer.QueName = "ApiServerExistSyncQueue-code-" + config.MachineCode
	consumer.AutoAck = false
	consumer.DeleteUnused = false
	consumer.Durable = true
	consumer.Exchange = "existSync"

	consumer.ConsumeAuto(func(msg amqp.Delivery) {
		if msg.Type == string(model.SyncInsert) {
			//不存在则更新
			if !global.ExistFilter.Lookup(msg.Body) {
				if !global.ExistFilter.Insert(msg.Body) {
					log.Printf("Sync exist filter of inserting hash-value %v error\n", msg)
					consumer.NackSafe(msg.DeliveryTag)
				}
			}
		} else if msg.Type == string(model.SyncDelete) {
			//存在则移除
			if global.ExistFilter.Lookup(msg.Body) {
				if !global.ExistFilter.Delete(msg.Body) {
					log.Printf("Sync exist filter of removing hash-value %v error\n", msg)
					consumer.NackSafe(msg.DeliveryTag)
				}
			}
		}
		consumer.AckOne(msg.DeliveryTag)
	}, 5*time.Second)
}
