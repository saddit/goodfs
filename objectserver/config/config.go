package config

import (
	"goodfs/lib/util/datasize"
	"gopkg.in/yaml.v3"
	"os"
	"time"
)

var (
	LocalAddr string
)

const (
	ConfFilePath = "conf/object-server.yaml"
)

type CacheConfig struct {
	MaxSizeMB     datasize.DataSize `yaml:"max-size-mb"`
	TTL           time.Duration     `yaml:"ttl"`
	CleanInterval time.Duration     `yaml:"clean-interval"`
	MaxItemSizeMB datasize.DataSize `yaml:"max-item-size-mb"`
}

type Config struct {
	Port         string        `yaml:"port"`
	StoragePath  string        `yaml:"storage-path"`
	TempPath     string        `yaml:"temp-path"`
	AmqpAddress  string        `yaml:"amqp-address"`
	BeatInterval time.Duration `yaml:"beat-interval"`
	Cache        CacheConfig   `yaml:"cache"`
}

func ReadConfig() Config {
	f, err := os.Open(ConfFilePath)
	if err != nil {
		panic(err)
	}
	var conf Config
	if err = yaml.NewDecoder(f).Decode(&conf); err != nil {
		panic(err)
	}
	return conf
}
