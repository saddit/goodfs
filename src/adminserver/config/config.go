package config

import (
	"common/etcd"
	"common/logs"
	"common/registry"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
)

const (
	ConfFilePath = "../../conf/admin-server.yaml"
)

type DiscoveryConfig struct {
	Group        string `yaml:"group" env:"GROUP" env-default:"goodfs"`
	DataServName string `yaml:"data-serv-name" env:"DATA_SERV_NAME" env-default:"objectserver"`
	MetaServName string `yaml:"meta-serv-name" env:"META_SERV_NAME" env-default:"metaserver"`
	ApiServName  string `yaml:"api-serv-name" env:"API_SERV_NAME" env-default:"apiserver"`
}

type Config struct {
	Port          string          `yaml:"port" env:"PORT" env-default:"80"`
	Log           logs.Config     `yaml:"log" env-prefix:"LOG"`
	Discovery     DiscoveryConfig `yaml:"discovery" env-prefix:"DISCOVERY"`
	Etcd          etcd.Config     `yaml:"etcd" env-prefix:"ETCD"`
	EnabledApiTLS bool            `yaml:"enabled-api-tls" env:"ENABLED_API_TLS"`
	registry      registry.Config `yaml:"registry" env-prefix:"REGISTRY"`
}

func (c *Config) init() {
	c.registry.Group = c.Discovery.Group
}

func (c *Config) GetRegistryCfg() *registry.Config {
	return &c.registry
}

func ReadConfig() Config {
	var conf Config
	if err := cleanenv.ReadConfig(ConfFilePath, &conf); err != nil {
		panic(err)
	}
	logs.Std().Infof("read config from %s", ConfFilePath)
	conf.init()
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
	conf.init()
	return conf
}
