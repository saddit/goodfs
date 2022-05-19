package config

import (
	"os"
	"time"

	yaml "gopkg.in/yaml.v3"
)

const (
	ConfFilePath  = "conf/api-server.yaml"
	MetadataMongo = "metadata_v2"
)

type Config struct {
	Port            string        `yaml:"port"`
	AmqpAddress     string        `yaml:"amqp-address"`
	MongoAddress    string        `yaml:"mongo-address"`
	LogDir          string        `yaml:"log-dir"`
	DetectInterval  time.Duration `yaml:"detect-interval"`
	SuspendTimeout  time.Duration `yaml:"suspend-timeout"`
	DeadTimeout     time.Duration `yaml:"dead-timeout"`
	SelectStrategy  string        `yaml:"select-strategy"`
	LocalCacheSize  string        `yaml:"local-cache-size"`
	EnableHashCheck bool          `yaml:"enable-hash-check"`
	Rs              RSConfig      `yaml:"rs"`
	//LocalStorePath  string        `yaml:"local-store-path"`
}

type RSConfig struct {
	DataShards    int `yaml:"data-shards"`
	ParityShards  int `yaml:"parity-shards"`
	BlockPerShard int `yaml:"block-per-shard"`
}

func (r *RSConfig) AllShards() int {
	return r.DataShards + r.ParityShards
}

func (r *RSConfig) BlockSize() int {
	return r.BlockPerShard * r.DataShards
}

func ReadConfig() Config {
	return ReadConfigFrom(ConfFilePath)
}

func ReadConfigFrom(path string) Config {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	var conf Config
	if err = yaml.NewDecoder(f).Decode(&conf); err != nil {
		panic(err)
	}
	return conf
}
