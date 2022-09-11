package config

import (
	"common/etcd"
	"common/logs"
	"common/registry"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	ConfFilePath = "../conf/api-server.yaml"
)

type Config struct {
	Port            string          `yaml:"port" env:"PORT" env-default:"8080"`
	LogLevel        logs.Level      `yaml:"log-level" env:"LOG_LEVEL"`
	SelectStrategy  string          `yaml:"select-strategy" env:"SELECT_STRATEGY" env-default:"random"`
	EnableHashCheck bool            `yaml:"enable-hash-check" env:"ENABLE_HASH_CHECK" env-default:"true"`
	Etcd            etcd.Config     `yaml:"etcd" env-prefix:"ETCD"`
	Rs              RsConfig        `yaml:"rs" env-prefix:"RS"`
	Discovery       DiscoveryConfig `yaml:"discovery" env-prefix:"DISCOVERY"`
	Registry        registry.Config `yaml:"registry" env-prefix:"REGISTRY"`
}

type DiscoveryConfig struct {
	DataServName string `yaml:"data-serv-name" env-default:"objectserver"`
	MetaServName string `yaml:"meta-serv-name" env-default:"metaserver"`
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
