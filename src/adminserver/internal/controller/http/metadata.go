package http

import (
	"adminserver/internal/usecase/logic"
	"adminserver/internal/usecase/pool"
	"adminserver/internal/usecase/webapi"
	"common/proto/msg"
	"common/response"
	"common/util"

	"github.com/gin-gonic/gin"
)

type MetadataController struct {
}

func NewMetadataController() *MetadataController {
	return &MetadataController{}
}

func (mc *MetadataController) Register(r gin.IRouter) {
	r.Group("metadata").Use(RequireToken).
		GET("/page", mc.Page).
		GET("/versions", mc.Versions).
		POST("/migration", mc.Migration).
		GET("/slots_detail", mc.SlotsDetail).
		GET("/peers", mc.Peers).
		GET("/buckets", mc.BucketList).
		POST("/create_bucket", mc.CreateBucket).
		PUT("/update_bucket", mc.UpdateBucket).
		DELETE("/delete_bucket/:name", mc.DeleteBucket).
		POST("/leave_cluster", mc.LeaveCluster).
		POST("/join_leader", mc.JoinLeader).
		GET("/config/:serverId", mc.GetConfig)
}

func (mc *MetadataController) Page(c *gin.Context) {
	var cond logic.MetadataCond
	if err := c.ShouldBindQuery(&cond); err != nil {
		response.FailErr(err, c)
		return
	}
	res, total, err := logic.NewMetadata().MetadataPaging(&cond)
	if err != nil {
		response.FailErr(err, c)
		return
	}
	response.Exec(c).
		Header(gin.H{"X-Total-Count": total}).
		JSON(res)
}

func (mc *MetadataController) Versions(c *gin.Context) {
	var cond logic.MetadataCond
	if err := c.ShouldBindQuery(&cond); err != nil {
		response.FailErr(err, c)
		return
	}
	res, total, err := logic.NewMetadata().VersionPaging(cond, GetAuthToken(c))
	if err != nil {
		response.FailErr(err, c)
		return
	}
	c.Header("X-Total-Count", util.IntString(total))
	if _, err := c.Writer.Write(res); err != nil {
		response.FailErr(err, c)
		return
	}
}

func (mc *MetadataController) Migration(c *gin.Context) {
	body := struct {
		SrcServerId  string   `json:"srcServerId" binding:"required"`
		DestServerId string   `json:"destServerId" binding:"required"`
		Slots        []string `json:"slots" binding:"required"`
	}{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailErr(err, c)
		return
	}
	if err := logic.NewMetadata().StartMigration(body.SrcServerId, body.DestServerId, body.Slots); err != nil {
		response.FailErr(err, c)
		return
	}
	response.Ok(c)
}

func (mc *MetadataController) SlotsDetail(c *gin.Context) {
	detail, err := logic.NewMetadata().GetSlotsDetail()
	if err != nil {
		response.FailErr(err, c)
		return
	}
	response.OkJson(detail, c)
}

func (mc *MetadataController) JoinLeader(c *gin.Context) {
	req := struct {
		ServerId string `json:"serverId" binding:"required"`
		MasterId string `json:"masterId" binding:"required"`
	}{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailErr(err, c)
		return
	}
	if err := logic.NewMetadata().JoinRaftCluster(req.MasterId, req.ServerId); err != nil {
		response.FailErr(err, c)
		return
	}
	response.Ok(c)
}

func (mc *MetadataController) LeaveCluster(c *gin.Context) {
	req := struct {
		ServerId string `json:"serverId" binding:"required"`
	}{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailErr(err, c)
		return
	}
	if err := logic.NewMetadata().LeaveRaftCluster(req.ServerId); err != nil {
		response.FailErr(err, c)
		return
	}
	response.Ok(c)
}

func (mc *MetadataController) Peers(c *gin.Context) {
	req := struct {
		ServerId string `form:"serverId" binding:"required"`
	}{}
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailErr(err, c)
		return
	}
	res, err := logic.NewMetadata().GetPeers(req.ServerId)
	if err != nil {
		response.FailErr(err, c)
		return
	}
	response.OkJson(res, c)
}

func (mc *MetadataController) BucketList(c *gin.Context) {
	var cond logic.BucketCond
	if err := c.ShouldBindQuery(&cond); err != nil {
		response.FailErr(err, c)
		return
	}
	list, total, err := logic.NewMetadata().BucketPaging(&cond)
	if err != nil {
		response.FailErr(err, c)
		return
	}
	response.Exec(c).
		Header(gin.H{"X-Total-Count": total}).
		JSON(list)
}

func (mc *MetadataController) GetConfig(c *gin.Context) {
	sid := c.Param("serverId")
	ip, ok := pool.Discovery.GetService(pool.Config.Discovery.MetaServName, sid)
	if !ok {
		response.BadRequestMsg("unknown serverId", c)
		return
	}
	jsonData, err := logic.NewMetadata().GetConfig(ip)
	if err != nil {
		response.FailErr(err, c)
		return
	}
	if _, err = c.Writer.Write(jsonData); err != nil {
		response.FailErr(err, c)
		return
	}
	response.Ok(c)
}

func (mc *MetadataController) CreateBucket(c *gin.Context) {
	var b msg.Bucket
	if err := c.ShouldBindJSON(&b); err != nil {
		response.FailErr(err, c)
		return
	}
	err := webapi.CreateBucket(logic.SelectApiServer(), &b, GetAuthToken(c))
	if err != nil {
		response.FailErr(err, c)
		return
	}
	response.Ok(c)
}

func (mc *MetadataController) UpdateBucket(c *gin.Context) {
	var b msg.Bucket
	if err := c.ShouldBindJSON(&b); err != nil {
		response.FailErr(err, c)
		return
	}
	err := webapi.UpdateBucket(logic.SelectApiServer(), &b, GetAuthToken(c))
	if err != nil {
		response.FailErr(err, c)
		return
	}
	response.Ok(c)
}

func (mc *MetadataController) DeleteBucket(c *gin.Context) {
	name := c.Param("name")
	err := webapi.DeleteBucket(logic.SelectApiServer(), name, GetAuthToken(c))
	if err != nil {
		response.FailErr(err, c)
		return
	}
	response.Ok(c)
}
