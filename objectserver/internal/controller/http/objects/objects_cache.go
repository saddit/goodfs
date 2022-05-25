package objects

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"objectserver/internal/entity"
	"objectserver/internal/usecase/pool"

	"github.com/gin-gonic/gin"
)

func GetFromCache(g *gin.Context) {
	name := g.Param("name")
	if bt, ok := pool.Cache.HasGet(name); ok {
		if _, e := g.Writer.Write(bt); e != nil {
			log.Printf("Match file cache %v, but written to response error: %v\n", name, e)
			g.AbortWithStatus(http.StatusInternalServerError)
		} else {
			log.Printf("Match file cache %v\n", name)
			g.AbortWithStatus(http.StatusOK)
		}
	} else {
		g.Writer = &entity.RespBodyWriter{Body: &bytes.Buffer{}, ResponseWriter: g.Writer}
	}
}

func getBody(g *gin.Context) (io.ReadCloser, error) {
	req := g.Request
	if req.Method == http.MethodPut {
		return req.GetBody()
	} else if req.Method == http.MethodGet {
		if w, ok := g.Writer.(*entity.RespBodyWriter); ok {
			return w, nil
		}
	}
	log.Panicf("Not support http method %v to save cache, check your route configuration", req.Method)
	return nil, nil
}

func SaveToCache(g *gin.Context) {
	name := g.Param("name")
	if pool.Cache.Has(name) {
		return
	}
	if body, e := getBody(g); e == nil {
		if g.Request.ContentLength < pool.Config.Cache.MaxItemSizeMB.Int64Value() {
			if bt, e := io.ReadAll(body); e == nil {
				pool.Cache.Set(name, bt)
				g.Set("Evict", false)
				log.Printf("Save %v to cache success\n", name)
			}
		} else {
			log.Printf("Skip too big cache: %v\n", name)
		}
	}
}

func RemoveCache(g *gin.Context) {
	name := g.Param("name")
	if evict, ok := g.Get("Evict"); ok && evict.(bool) {
		pool.Cache.Delete(name)
		log.Printf("Success evict cache %v\n", name)
	}
}
