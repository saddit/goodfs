package stat

import (
	"common/response"
	"common/system/disk"
	"net/http"
	"objectserver/internal/usecase/pool"

	"github.com/gin-gonic/gin"
)

func Ping(c *gin.Context) {
	c.Status(http.StatusOK)
}

func StatInfo(c *gin.Context) {
	hd := gin.H{
		"Capacity": pool.ObjectCap.Capacity(),
	}
	if info, err := disk.GetAverageIOStats(); err == nil {
		hd["Weighted-IO"] = info.WeightedIO
		hd["IO-Time"] = info.IoTime
	}
	response.OkHeader(hd, c)
}