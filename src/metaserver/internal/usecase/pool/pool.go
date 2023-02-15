package pool

import (
	"common/cache"
	"common/cst"
	"common/datasize"
	"common/etcd"
	"common/logs"
	"common/registry"
	"common/util"
	"fmt"
	"metaserver/config"
	"metaserver/internal/usecase/db"
	"metaserver/internal/usecase/raftimpl"
	"path/filepath"
	"time"

	"github.com/allegro/bigcache"
	"github.com/gin-gonic/gin"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	Config       *config.Config
	Cache        cache.ICache
	Storage      *db.Storage
	HashSlot     *db.HashSlotDB
	RaftWrapper  *raftimpl.RaftWrapper
	Etcd         *clientv3.Client
	Registry     *registry.EtcdRegistry
	HttpHostPort string
	GrpcHostPort string
)

func InitPool(cfg *config.Config) {
	Config = cfg
	HttpHostPort = util.ServerAddress(cfg.Port)
	GrpcHostPort = util.ServerAddress(cfg.Cluster.Port)
	initLog(&cfg.Log)
	initCache(cfg.Cache)
	initEtcd(&cfg.Etcd)
	initRegistry(&cfg.Registry, Etcd)
	initStorage(cfg)
	initHashSlot(&cfg.Registry, Etcd)
}

func initLog(cfg *logs.Config) {
	logs.SetLevel(cfg.Level)
	if logs.IsDebug() || logs.IsTrace() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
}

func initEtcd(cfg *etcd.Config) {
	var err error
	Etcd, err = clientv3.New(clientv3.Config{
		Endpoints:           cfg.Endpoint,
		Username:            cfg.Username,
		Password:            cfg.Password,
		PermitWithoutStream: true,
	})
	if err != nil {
		panic(fmt.Errorf("create etcd client err: %v", err))
	}
}

func initStorage(cfg *config.Config) {
	// open db file
	Storage = db.NewStorage()
	if err := Storage.Open(filepath.Join(cfg.DataPath, cfg.Registry.ServerID+".db")); err != nil {
		panic(fmt.Errorf("open db err: %v", err))
	}
}

func initRegistry(cfg *registry.Config, etcd *clientv3.Client) {
	cfg.HttpAddr = HttpHostPort
	cfg.RpcAddr = GrpcHostPort
	Registry = registry.NewEtcdRegistry(etcd, *cfg)
}

func initHashSlot(cfg *registry.Config, etcd *clientv3.Client) {
	HashSlot = db.NewHashSlotDB(cst.EtcdPrefix.FmtHashSlot(cfg.Group, cfg.Name, ""), etcd)
}

func initCache(cfg config.CacheConfig) {
	conf := bigcache.DefaultConfig(cfg.TTL)
	conf.HardMaxCacheSize = int(cfg.MaxSize.MegaByte())
	conf.CleanWindow = cfg.CleanInterval
	conf.Verbose = false
	conf.Shards = 2048
	conf.MaxEntrySize = int(datasize.KB * 4)
	conf.MaxEntriesInWindow = int(cfg.MaxSize / (8 * datasize.KB))
	Cache = cache.NewCache(conf)
}

func Close() {
	util.LogErr(Storage.Stop())
	util.LogErr(Etcd.Close())
	util.LogErr(HashSlot.Close(time.Minute))
	if RaftWrapper != nil {
		util.LogErr(RaftWrapper.Close())
	}
}
