package pool

import (
	"apiserver/config"
	"apiserver/internal/usecase/componet/selector"
	"common/logs"
	"common/performance"
	"common/registry"
	"common/util"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	Config    *config.Config
	Etcd      *clientv3.Client
	Http      *http.Client
	Balancer  selector.Selector
	Discovery *registry.EtcdDiscovery
	Perform   performance.Collector
)

func InitPool(cfg *config.Config) {
	Config = cfg
	initLog(&cfg.Log)
	initHttp()
	initEtcd(cfg)
	initDiscovery(Etcd, cfg)
	initBalancer(cfg)
	initPerform(&cfg.Performance, &cfg.Log, &cfg.Registry, Etcd)
}

func Close() {
	Http.CloseIdleConnections()
	util.LogErr(Perform.Close())
	util.LogErr(Etcd.Close())
}

func initEtcd(cfg *config.Config) {
	// init etcd
	var err error
	Etcd, err = clientv3.New(clientv3.Config{
		Endpoints:           cfg.Etcd.Endpoint,
		Username:            cfg.Etcd.Username,
		Password:            cfg.Etcd.Password,
		PermitWithoutStream: true,
	})
	if err != nil {
		panic(err)
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

func initHttp() {
	Http = &http.Client{Timeout: 5 * time.Second}
}

func initBalancer(cfg *config.Config) {
	Balancer = selector.NewSelector(cfg.SelectStrategy)
}

func initPerform(cfg *performance.Config, logCfg *logs.Config, regCfg *registry.Config, etcd *clientv3.Client) {
	if cfg.Store == performance.Local {
		localPath := logCfg.StoreDir
		if localPath == "" {
			localPath = os.TempDir()
		}
		performance.SetLocalStore(performance.NewLocalStore(filepath.Join(localPath, regCfg.ServerID+".perf")))
	}
	if cfg.Store == performance.Remote {
		performance.SetRemoteStore(performance.NewEtcdStore(etcd, []string{
			performance.ActionRead,
			performance.ActionWrite,
		}))
	}
	Perform = performance.NewCollector(cfg)
}
