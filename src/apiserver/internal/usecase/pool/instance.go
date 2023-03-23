package pool

import (
	"apiserver/config"
	"apiserver/internal/usecase/componet/selector"
	"apiserver/internal/usecase/grpcapi"
	"apiserver/internal/usecase/webapi"
	"common/logs"
	"common/performance"
	"common/registry"
	"common/util"
	"github.com/gin-gonic/gin"
	clientv3 "go.etcd.io/etcd/client/v3"
	"os"
	"path/filepath"
	"time"
)

var (
	Config    *config.Config
	Etcd      *clientv3.Client
	Balancer  selector.Selector
	Discovery *registry.EtcdDiscovery
	Perform   performance.Collector
)

func InitPool(cfg *config.Config) {
	Config = cfg
	initLog(&cfg.Log)
	initEtcd(cfg)
	initDiscovery(Etcd, cfg)
	initBalancer(cfg)
	initPerform(&cfg.Performance, &cfg.Log, &cfg.Registry, Etcd)
}

func Close() {
	util.LogErr(Perform.Close())
	util.LogErr(Etcd.Close())
	util.LogErr(grpcapi.Close())
	webapi.Close()
}

func initEtcd(cfg *config.Config) {
	// init etcd
	var err error
	Etcd, err = clientv3.New(clientv3.Config{
		DialTimeout:         10 * time.Second,
		Endpoints:           cfg.Etcd.Endpoint,
		Username:            cfg.Etcd.Username,
		Password:            cfg.Etcd.Password,
		PermitWithoutStream: true,
	})
	if err != nil {
		panic("init etcd fail: " + err.Error())
	}
}

func initDiscovery(etcd *clientv3.Client, cfg *config.Config) {
	cfg.Registry.Services = []string{cfg.Discovery.DataServName, cfg.Discovery.MetaServName}
	Discovery = registry.NewEtcdDiscovery(etcd, &cfg.Registry)
}

func initLog(cfg *logs.Config) {
	logs.SetLevel(cfg.Level)
	if logs.IsDebug() || logs.IsTrace() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
}

func initBalancer(cfg *config.Config) {
	Balancer = selector.NewSelector(cfg.SelectStrategy)
}

func initPerform(cfg *performance.Config, logCfg *logs.Config, regCfg *registry.Config, etcd *clientv3.Client) {
	if cfg.Enable && cfg.Store == performance.Local {
		localPath := logCfg.StoreDir
		if localPath == "" {
			localPath = os.TempDir()
		}
		performance.SetLocalStore(performance.NewLocalStore(filepath.Join(localPath, regCfg.SID()+".perf")))
	}
	if cfg.Enable && cfg.Store == performance.Remote {
		performance.SetRemoteStore(performance.NewEtcdStore(etcd, []string{
			performance.ActionRead,
			performance.ActionWrite,
		}))
	}
	Perform = performance.NewCollector(cfg)
	webapi.SetPerformanceCollector(Perform)
	grpcapi.SetPerformanceCollector(Perform)
}
