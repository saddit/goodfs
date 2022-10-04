package config

import (
	"common/datasize"
	"common/etcd"
	"common/logs"
	"common/registry"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	ConfFilePath = "../conf/object-server.yaml"
)

type CacheConfig struct {
	TTL           time.Duration     `yaml:"ttl" env:"TTL" env-default:"1h"`
	CleanInterval time.Duration     `yaml:"clean-interval" env:"CLEAN_INTERVAL" env-default:"1h"`
	MaxItemSize   datasize.DataSize `yaml:"max-item-size" env:"MAX_ITEM_SIZE" env-default:"12MB"`
	MaxSize       datasize.DataSize `yaml:"max-size" env:"MAX_SIZE" env-default:"128MB"`
}

type Config struct {
	Port        string          `yaml:"port" env-default:"8100"`
	ServerID    string          `yaml:"server-id" env-required:"true"`
	RpcPort     string          `yaml:"rpc-port" env-default:"4100"`
	LogLevel    logs.Level      `yaml:"log-level" env:"LOG_LEVEL" env-default:"INFO"`
	StoragePath string          `yaml:"storage-path" env:"STORAGE_PATH" env-default:"objects"`
	TempPath    string          `yaml:"temp-path" env:"TEMP_PATH" env-default:"temp"`
	Cache       CacheConfig     `yaml:"cache" env-prefix:"CACHE"`
	Etcd        etcd.Config     `yaml:"etcd" env-prefix:"ETCD"`
	Registry    registry.Config `yaml:"registry" env-prefix:"REGISTRY"`
}

func ReadConfig() Config {
	var conf Config
	if err := cleanenv.ReadConfig(ConfFilePath, &conf); err != nil {
		panic(err)
	}
	logs.Std().Infof("read config from %s", ConfFilePath)
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
	return conf
}
