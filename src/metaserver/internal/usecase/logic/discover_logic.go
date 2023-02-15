package logic

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"metaserver/internal/usecase/pool"
	"net/url"
)

type Discover struct {
}

func NewDiscovery() Discover {
	return Discover{}
}

func (Discover) PeerIp(id string, rpc bool) string {
	ip, _ := pool.Registry.GetService(pool.Config.Registry.Name, id, rpc)
	return ip
}

func (d Discover) PeerLocation(id string, c *gin.Context) string {
	ip := d.PeerIp(id, false)
	if ip == "" {
		ip = "unknown-id"
	}
	loc, _ := url.JoinPath(fmt.Sprint("http://", ip), c.Request.URL.RawPath)
	return loc
}
