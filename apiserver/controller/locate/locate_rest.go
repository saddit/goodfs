package locate

import (
	"goodfs/apiserver/config"
	"goodfs/apiserver/global"
	"goodfs/apiserver/model"
	"goodfs/apiserver/service"
	"goodfs/util"
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
	consumer.DeleteUnused = false
	consumer.Durable = true
	consumer.Exchange = "existSync"
	consumeChan, ok := consumer.Consume()

	//断线重连策略
	for range util.ImmediateTick(5 * time.Second) {
		if ok {
			log.Println("Start syncing existing hash-value!")
			for msg := range consumeChan {
				if msg.Type == string(model.SyncInsert) {
					//不存在则更新
					if !global.ExistFilter.Lookup(msg.Body) {
						if !global.ExistFilter.Insert(msg.Body) {
							log.Printf("Sync exist filter of inserting hash-value %v error\n", msg)
						}
					}
				} else if msg.Type == string(model.SyncDelete) {
					if !global.ExistFilter.Delete(msg.Body) {
						log.Printf("Sync exist filter of removing hash-value %v error\n", msg)
					}
				}
			}
			ok = false
		} else {
			log.Println("Recovering syncing existing hash-value consumer...")
			//重试直到成功
			consumeChan, ok = consumer.Consume()
		}
	}
}
