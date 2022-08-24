package config

import (
	"common/etcd"
	"common/logs"
	"common/registry"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	ConfFilePath = "../conf/meta-server.yaml"
)

type Config struct {
	Port     string          `yaml:"port" env:"PORT" env-default:"4091"`
	LogLevel logs.Level      `yaml:"log-level" env:"LOG_LEVEL"`
	DataDir  string          `yaml:"data-dir" env:"DATA_DIR" env-default:"/tmp/goodfs"`
	Cluster  ClusterConfig   `yaml:"cluster" env-prefix:"CLUSTER"`
	Registry registry.Config `yaml:"registry" env-prefix:"REGISTRY"`
	Etcd     etcd.Config     `yaml:"etcd" env-prefix:"ETCD"`
}

type ClusterConfig struct {
	ID               string        `yaml:"id" env:"ID" env-required:"true"`
	Port             string        `yaml:"port" env:"PORT" env-default:"4092"`
	Bootstrap        bool          `yaml:"bootstrap" env:"BOOTSTRAP" env-default:"false"`
	ElectionTimeout  time.Duration `yaml:"election-timeout" env:"ELECTION_TIMEOUT" env-default:"900ms"`
	HeartbeatTimeout time.Duration `yaml:"heartbeat-timeout" env:"HEARTBEAT_TIMEOUT" env-default:"800ms"`
	LogLevel         string        `yaml:"log-level" env:"LOG_LEVEL" env-default:"INFO"`
	StoreDir         string        `yaml:"store-dir" env:"STORE_DIR" env-default:"/tmp/goodfs_metaserver"`
	Nodes            []string      `yaml:"nodes" env:"NODES" env-separator:","`
}

func ReadConfig() Config {
	var conf Config
	if err := cleanenv.ReadConfig(ConfFilePath, &conf); err != nil {
		panic(err)
	}
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
	return conf
}
