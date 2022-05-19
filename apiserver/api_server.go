package main

import (
	"fmt"
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
	"goodfs/apiserver/command"
	"goodfs/apiserver/config"
	"goodfs/apiserver/controller"
	"goodfs/apiserver/controller/heartbeat"
	"goodfs/apiserver/global"
	"goodfs/apiserver/service/selector"
	"goodfs/lib/graceful"
	"log"
	"net/http"
	"time"

	"github.com/838239178/goodmq"
	"github.com/gin-gonic/gin"
)

func initialize() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&nested.Formatter{
		HideKeys:    true,
		FieldsOrder: []string{"component", "category"},
	})
	global.Config = config.ReadConfig()
	global.Http = &http.Client{Timeout: 5 * time.Second}
	goodmq.RecoverDelay = 3 * time.Second
	global.AmqpConnection = goodmq.NewAmqpConnection(global.Config.AmqpAddress)
	global.Balancer = selector.NewSelector(global.Config.SelectStrategy)
	//var e error
	//if global.LocalDB, e = leveldb.OpenFile(global.Config.LocalStorePath, &opt.Options{
	//	BlockCacheCapacity:          datasize.MustParse(global.Config.LocalCacheSize).IntValue(),
	//	CompactionSourceLimitFactor: 5,
	//}); e != nil {
	//	panic(e)
	//}

	command.ReadCommand()
}

func shutdown() {
	err := global.AmqpConnection.Close()
	if err != nil {
		log.Println(err)
	}
	global.Http.CloseIdleConnections()
	//if err = global.LocalDB.Close(); err != nil {
	//	log.Println(err)
	//}
}

func main() {
	initialize()
	defer shutdown()

	go heartbeat.ListenHeartbeat()

	router := gin.Default()

	api := router.Group("/api")
	controller.Router(api)

	graceful.ListenAndServe(fmt.Sprint(":", global.Config.Port), router)
}
