package logic

import (
	"adminserver/internal/usecase/pool"
	"math/rand"
	"time"
)

func SelectApiServer() string {
	servers := pool.Discovery.GetServices(pool.Config.Discovery.ApiServName, false)
	rand.Seed(time.Now().Unix())
	idx := rand.Intn(len(servers))
	return servers[idx]
}
