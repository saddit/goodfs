package objects

import (
	"fmt"
	"goodfs/objects/config"
	"goodfs/objects/heartbeat"
	"goodfs/objects/locate"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

func initalize() {
	hn, e := os.Hostname()
	if e != nil {
		panic(e)
	}
	config.LocalAddr = fmt.Sprintf("%v:%v", hn, config.Port)
}

func Start() {
	initalize()

	go heartbeat.StartHeartbeat()
	go locate.StartLocate()

	router := gin.Default()

	//init router
	Router(router)

	router.Run(":" + strconv.Itoa(config.Port))
}
