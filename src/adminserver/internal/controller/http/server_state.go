package http

import (
	"adminserver/internal/entity"
	"adminserver/internal/usecase/logic"
	"adminserver/internal/usecase/pool"
	"common/response"
	"common/util"

	"github.com/gin-gonic/gin"
)

type ServerStateController struct {
	logic logic.ServerMonitor
}

func NewServerStateController() *ServerStateController {
	return &ServerStateController{}
}

func (ss *ServerStateController) Register(r gin.IRouter) {
	r.Group("server").
		GET("/stat", ss.Stat).
		GET("/overview", ss.Overview).
		GET("/etcdstat", ss.EtcdStat).
		GET("/:type/timeline", ss.UsageTimeline)
}

func (ss *ServerStateController) Overview(c *gin.Context) {
	aliveCounts := ss.logic.AliveCounts()
	metaLogic := logic.NewMetadata()
	nb, err := metaLogic.TotalBuckets()
	if err != nil {
		response.FailErr(err, c)
		return
	}
	no, err := metaLogic.TotalObjects()
	if err != nil {
		response.FailErr(err, c)
		return
	}
	cpu, mem := ss.logic.StatTimeLineOverview()
	response.OkJson(gin.H{
		"aliveCounts":  aliveCounts,
		"totalBuckets": nb,
		"totalObjects": no,
		"avgCpu":       cpu,
		"avgMem":       mem,
	}, c)
}

func (ss *ServerStateController) Stat(c *gin.Context) {
	monitor := logic.NewServerMonitor()
	dg := util.NewDoneGroup()
	dg.Todo()
	var info [3]map[string]*entity.ServerInfo
	go func() {
		defer dg.Done()
		metaInfo, err := monitor.ServerStat(pool.Config.Discovery.MetaServName)
		if err != nil {
			dg.Error(err)
			return
		}
		// mark master server
		masters := logic.NewMetadata().GetMasterServerIds()
		for _, v := range metaInfo {
			v.IsMaster = masters.Contains(v.ServerID)
		}
		info[0] = metaInfo
	}()
	dg.Todo()
	go func() {
		defer dg.Done()
		dataInfo, err := monitor.ServerStat(pool.Config.Discovery.DataServName)
		if err != nil {
			dg.Error(err)
			return
		}
		info[1] = dataInfo
	}()
	dg.Todo()
	go func() {
		defer dg.Done()
		apiInfo, err := monitor.ServerStat(pool.Config.Discovery.ApiServName)
		if err != nil {
			dg.Error(err)
			return
		}
		info[2] = apiInfo
	}()
	if err := dg.WaitUntilError(); err != nil {
		response.FailErr(err, c)
		return
	}
	response.OkJson(gin.H{
		"metaServer": info[0],
		"dataServer": info[1],
		"apiServer":  info[2],
	}, c)
}

func (ss *ServerStateController) UsageTimeline(c *gin.Context) {
	usageType := c.Param("type")
	sn := c.Query("server")
	res := logic.NewServerMonitor().StatTimeline(util.ToInt(sn), usageType)
	response.OkJson(res, c)
}

func (ss *ServerStateController) EtcdStat(c *gin.Context) {
	res, err := ss.logic.EtcdStatus()
	if err != nil {
		response.FailErr(err, c)
		return
	}
	response.OkJson(res, c)
}
