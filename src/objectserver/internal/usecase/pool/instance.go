package pool

import (
	"common/cache"
	"common/collection/set"
	"common/cst"
	"common/datasize"
	"common/etcd"
	"common/graceful"
	"common/logs"
	"common/registry"
	"common/util"
	"common/util/slices"
	"errors"
	"objectserver/config"
	"objectserver/internal/db"
	"objectserver/internal/usecase/component"
	"os"
	"path/filepath"
	"sync"

	"github.com/allegro/bigcache"
	"github.com/gin-gonic/gin"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	Config        *config.Config
	Etcd          *clientv3.Client
	ObjectCap     *db.ObjectCapacity
	PathDB        *db.PathCache
	DriverManager *component.DriverManager
	Cache         cache.ICache
	Registry      registry.Registry
	Discovery     registry.Discovery
)

var (
	openFn       func()
	onCloseEvent []func()
	closeOnce    = &sync.Once{}
	openOnce     = &sync.Once{}
)

// OnClose as defer on pool.Close(). Last in first invoke.
func OnClose(fn ...func()) {
	onCloseEvent = append(onCloseEvent, fn...)
}

func OnOpen(fn func()) {
	openFn = fn
}

func Open() {
	openOnce.Do(func() {
		closeOnce = &sync.Once{}
		openFn()
	})
}

func OpenGraceful() (err error) {
	defer graceful.Recover(func(msg string) {
		err = errors.New(msg)
	})
	Open()
	return
}

func CloseGraceful() (err error) {
	defer graceful.Recover(func(msg string) {
		err = errors.New(msg)
	})
	Close()
	return
}

func InitPool(cfg *config.Config) {
	Config = cfg
	initDriverManger(cfg.ExcludeMountPoints, cfg.AllowedMountPoints)
	initDir(cfg, DriverManager)
	initLog(&cfg.Log)
	initCache(&cfg.Cache)
	initEtcd(&cfg.Etcd)
	initRegister(Etcd, cfg)
	initObjectCap(Etcd, cfg)
	initPathCache(cfg)
}

func initDir(cfg *config.Config, dm *component.DriverManager) {
	if e := os.MkdirAll(filepath.Join(cfg.BaseMountPoint, cfg.PathCachePath), cst.OS.ModeUser); e != nil {
		panic(e)
	}
	dm.Update()
	for _, mp := range dm.GetAllMountPoint() {
		if e := os.MkdirAll(filepath.Join(mp, cfg.TempPath), cst.OS.ModeUser); e != nil {
			panic(e)
		}
		if e := os.MkdirAll(filepath.Join(mp, cfg.StoragePath), cst.OS.ModeUser); e != nil {
			panic(e)
		}
	}
}

func initDriverManger(ex []string, in []string) {
	DriverManager = component.NewDriverManager(component.SpaceFirstBalancer())
	DriverManager.Excludes = set.OfString(ex)
	DriverManager.Includes = set.OfString(in)
}

func initPathCache(cfg *config.Config) {
	var err error
	PathDB, err = db.NewPathCache(filepath.Join(cfg.BaseMountPoint, cfg.PathCachePath))
	if err != nil {
		panic(err)
	}
}

func initLog(cfg *logs.Config) {
	logs.SetLevel(cfg.Level)
	if logs.IsDebug() || logs.IsTrace() {
		_ = os.Setenv(util.ServerIpEnv, "127.0.0.1")
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
}

func initObjectCap(et *clientv3.Client, cfg *config.Config) {
	ObjectCap = db.NewObjectCapacity(et, cfg)
}

func initCache(cfg *config.CacheConfig) {
	cacheConf := bigcache.DefaultConfig(cfg.TTL)
	cacheConf.CleanWindow = cfg.CleanInterval
	cacheConf.HardMaxCacheSize = int(cfg.MaxSize.MegaByte())
	cacheConf.MaxEntrySize = int(datasize.KB * 4)
	cacheConf.Shards = 2048
	cacheConf.Verbose = false
	cacheConf.MaxEntriesInWindow = int(cfg.MaxSize / cfg.MaxItemSize)
	Cache = cache.NewCache(cacheConf)
}

func initEtcd(cfg *etcd.Config) {
	var e error
	if Etcd, e = clientv3.New(clientv3.Config{
		Endpoints:           cfg.Endpoint,
		Username:            cfg.Username,
		Password:            cfg.Password,
		PermitWithoutStream: true,
	}); e != nil {
		panic(e)
	}
}

func initRegister(et *clientv3.Client, cfg *config.Config) {
	cfg.Registry.HttpAddr = util.ServerAddress(cfg.Port)
	cfg.Registry.RpcAddr = util.ServerAddress(cfg.RpcPort)
	er := registry.NewEtcdRegistry(et, cfg.Registry)
	Registry, Discovery = er, er
}

func CloseAll() {
	defer Etcd.Close()
	defer Cache.Close()
	defer PathDB.Close()
	defer Close()
}

func Close() {
	closeOnce.Do(func() {
		defer slices.Clear(&onCloseEvent)
		openOnce = &sync.Once{}
		for _, fn := range onCloseEvent {
			//goland:noinspection GoDeferInLoop
			defer fn()
		}
	})
}
