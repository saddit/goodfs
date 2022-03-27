package objects

import (
	"goodfs/objects/config"
	"goodfs/objects/heartbeat"
	"log"
	"net/http"
	"strconv"
)

func Start() {
	go heartbeat.StartHeartbeat()
	http.HandleFunc("/objects/", Handler)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.Port), nil))
}
