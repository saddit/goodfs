package config

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	ConfFilePath  = "../conf/api-server.yaml"
	MetadataMongo = "metadata_v2"
)

type Config struct {
	Port            string          `yaml:"port" env:"PORT" env-default:"8080"`
	AmqpAddress     string          `yaml:"amqp-address" env:"AMQP_ADDRESS" env-required:"true"`
	LogDir          string          `yaml:"log-dir" env:"LOG_DIR" env-default:"logs"`
	SelectStrategy  string          `yaml:"select-strategy" env:"SELECT_STRATEGY" env-default:"random"`
	EnableHashCheck bool            `yaml:"enable-hash-check" env:"ENABLE_HASH_CHECK" env-default:"true"`
	Etcd            EtcdConfig      `yaml:"etcd" env-prefix:"ETCD"`
	Rs              RsConfig        `yaml:"rs" env-prefix:"RS"`
	Discovery       DiscoveryConfig `yaml:"discovery" env-prefix:"DISCOVERY"`
	Registry        RegistryConfig  `yaml:"registry" env-prefix:"REGISTRY"`
}

type EtcdConfig struct {
	Endpoint []string `yaml:"endpoint" env:"ENDPOINT" env-required:"true" env-separator:","`
	Username string   `yaml:"username" env:"USERNAME" env-required:"true"`
	Password string   `yaml:"password" env:"PASSWORD" env-required:"true"`
}

type RegistryConfig struct {
	Group    string        `yaml:"group" env:"GROUP" env-default:"goodfs"`
	Name     string        `yaml:"name" env:"NAME" env-default:"apiserver"`
	Interval time.Duration `yaml:"interval" env:"INTERVAL" env-default:"5s"`
	Timeout  time.Duration `yaml:"timeout" env:"TIMEOUT" env-default:"3s"`
}

type DiscoveryConfig struct {
	DetectInterval time.Duration `yaml:"detect-interval" env:"DETECT_INTERVAL" env-default:"5s"`
	SuspendTimeout time.Duration `yaml:"suspend-timeout" env:"SUSPEND_TIMEOUT" env-default:"5s"`
	DeadTimeout    time.Duration `yaml:"dead-timeout" env:"DEAD_TIMEOUT" env-default:"10s"`
}

type RsConfig struct {
	DataShards    int `yaml:"data-shards" env:"DATA_SHARDS" env-default:"4"`
	ParityShards  int `yaml:"parity-shards" env:"PARITY_SHARDS" env-default:"2"`
	BlockPerShard int `yaml:"block-per-shard" env:"BLOCK_PER_SHARD" env-default:"8000"`
}

func (r *RsConfig) AllShards() int {
	return r.DataShards + r.ParityShards
}

func (r *RsConfig) BlockSize() int {
	return r.BlockPerShard * r.DataShards
}

func ReadConfig() Config {
	return ReadConfigFrom(ConfFilePath)
}

func ReadConfigFrom(path string) Config {
	var conf Config
	if err := cleanenv.ReadConfig(path, &conf); err != nil {
		panic(err)
	}
	return conf
}
