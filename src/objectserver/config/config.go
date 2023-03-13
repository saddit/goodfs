package config

import (
	"common/datasize"
	"common/etcd"
	"common/logs"
	"common/registry"
	"os"
	"path/filepath"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	ConfFilePath = "../../conf/object-server.yaml"
)

type CacheConfig struct {
	TTL           time.Duration     `yaml:"ttl" env:"TTL" env-default:"1h"`
	CleanInterval time.Duration     `yaml:"clean-interval" env:"CLEAN_INTERVAL" env-default:"1h"`
	MaxItemSize   datasize.DataSize `yaml:"max-item-size" env:"MAX_ITEM_SIZE" env-default:"12MB"`
	MaxSize       datasize.DataSize `yaml:"max-size" env:"MAX_SIZE" env-default:"128MB"`
}

type StateConfig struct {
	SyncInterval time.Duration `yaml:"sync-interval" env:"SYNC_INTERVAL" env-default:"1m"`
}

type DiscoveryConfig struct {
	MetaServName string `yaml:"meta-serv-name" env-default:"metaserver"`
}

type innerConf struct {
	PathCachePath string `yaml:"-" env:"-"` // PathCachePath is a path to store path-db-file under BaseMountPoint
	TempPath      string `yaml:"-" env:"-"` // TempPath is a path to store temporary object file under different mount points
}

type Config struct {
	innerConf
	Port               string          `yaml:"port" env-default:"8100"`                                           // Port is port which the http server will listen to
	BaseMountPoint     string          `yaml:"base-mount-point" env:"BASE_MOUNT_POINT" env-required:"true"`       // BaseMountPoint refers a mount point to store central data also as a fallback choice.
	StoragePath        string          `yaml:"storage-path" env:"STORAGE_PATH" env-default:"/objects"`            // StoragePath is a path to store object file under different mount points
	AllowedMountPoints []string        `yaml:"allowed-mount-points" env:"ALLOWED_MOUNT_POINTS" env-separator:","` // AllowedMountPoints limits only these mount points allowed to store object file. Priority over ExcludeMountPoints but not affect BaseMountPoint.
	ExcludeMountPoints []string        `yaml:"exclude-mount-points" env:"EXCLUDE_MOUNT_POINTS" env-separator:","` // ExcludeMountPoints avoids to store object file under these mount points
	TempCleaners       int             `yaml:"temp-cleaners" env:"TEMP_CLEANERS" env-default:"3"`
	Log                logs.Config     `yaml:"log" env-prefix:"LOG"`
	State              StateConfig     `yaml:"state" env-prefix:"STATE"`
	Cache              CacheConfig     `yaml:"cache" env-prefix:"CACHE"`
	Etcd               etcd.Config     `yaml:"etcd" env-prefix:"ETCD"`
	Registry           registry.Config `yaml:"registry" env-prefix:"REGISTRY"`
	Discovery          DiscoveryConfig `yaml:"discovery" env-prefix:"DISCOVERY"`
}

func (c *Config) initialize() {
	c.PathCachePath = filepath.Join(c.StoragePath, c.Registry.ServerID+"_path-cache")
	c.TempPath = filepath.Join(c.StoragePath, c.Registry.ServerID+"_temp")
	c.StoragePath = filepath.Join(c.StoragePath, c.Registry.ServerID+"_store")
}

func ReadConfig() Config {
	var conf Config
	if err := cleanenv.ReadConfig(ConfFilePath, &conf); err != nil {
		panic(err)
	}
	logs.Std().Infof("read config from %s", ConfFilePath)
	conf.initialize()
	return conf
}

func ReadConfigFrom(path string) Config {
	var conf Config
	if err := cleanenv.ReadConfig(path, &conf); err != nil {
		if os.IsNotExist(err) {
			return ReadConfig()
		}
		panic(err)
	}
	logs.Std().Infof("read config from %s", path)
	conf.initialize()
	return conf
}
