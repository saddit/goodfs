package config

import (
	"common/cst"
	"common/datasize"
	"common/etcd"
	"common/logs"
	"common/registry"
	"os"
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

type Config struct {
	Port        string          `yaml:"port" env-default:"8100"`
	RpcPort     string          `yaml:"rpc-port" env-default:"4100"`
	StoragePath string          `yaml:"storage-path" env:"STORAGE_PATH" env-default:"objects"`
	TempPath    string          `yaml:"temp-path" env:"TEMP_PATH" env-default:"temp"`
	Log         logs.Config     `yaml:"log" env-prefix:"LOG"`
	State       StateConfig     `yaml:"state" env-prefix:"STATE"`
	Cache       CacheConfig     `yaml:"cache" env-prefix:"CACHE"`
	Etcd        etcd.Config     `yaml:"etcd" env-prefix:"ETCD"`
	Registry    registry.Config `yaml:"registry" env-prefix:"REGISTRY"`
	Discovery   DiscoveryConfig `yaml:"discovery" env-prefix:"DISCOVERY"`
}

func (c *Config) initialize() {
	if e := os.MkdirAll(c.TempPath, cst.OS.ModeUser); e != nil {
		panic(e)
	}
	if e := os.MkdirAll(c.StoragePath, cst.OS.ModeUser); e != nil {
		panic(e)
	}
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
