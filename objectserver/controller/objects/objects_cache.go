package objects

import (
	"goodfs/objectserver/config"
	"goodfs/objectserver/global"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetFromCache(g *gin.Context) {
	name := g.Param("name")
	if bt, ok := global.Cache.HasGet(name); ok {
		if _, e := g.Writer.Write(bt); e != nil {
			log.Printf("Match file cache %v, but written to response error: %v\n", name, e)
			g.AbortWithStatus(http.StatusInternalServerError)
		} else {
			log.Printf("Match file cache %v\n", name)
			g.AbortWithStatus(http.StatusOK)
		}
	}
}

func getBody(req *http.Request) (io.ReadCloser, error) {
	if req.Method == http.MethodPut {
		return req.GetBody()
	} else if req.Method == http.MethodGet {
		return req.Response.Body, nil
	}
	log.Panicf("Not support http method %v to save cache, check your route configuration", req.Method)
	return nil, nil
}

func SaveToCache(g *gin.Context) {
	name := g.Param("name")
	if body, e := getBody(g.Request); e == nil {
		if g.Request.ContentLength < config.CacheItemMaxSize.Int64Value() {
			bt := make([]byte, 0, g.Request.ContentLength)
			if _, e = body.Read(bt); e == nil {
				global.Cache.Set(name, bt)
				g.Keys["Evict"] = false
				log.Printf("Save %v to cache success\n", name)
			}
		} else {
			log.Printf("Skip too big cache: %v\n", name)
		}
	}
}

func RemoveCache(g *gin.Context) {
	name := g.Param("name")
	if evict := g.Keys["Evict"].(bool); evict {
		global.Cache.Delete(name)
		log.Printf("Success evict cache %v\n", name)
	}
}
