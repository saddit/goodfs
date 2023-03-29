package logic

import (
	"adminserver/internal/usecase/pool"
	"adminserver/internal/usecase/webapi"
	"math/rand"
	"time"
)

func SelectApiServer() string {
	servers := pool.Discovery.GetServices(pool.Config.Discovery.ApiServName)
	rand.Seed(time.Now().Unix())
	idx := rand.Intn(len(servers))
	return servers[idx]
}

func GetAPIConfig(ip, token string) ([]byte, error) {
	return webapi.GetAPIConfig(ip, token)
}
