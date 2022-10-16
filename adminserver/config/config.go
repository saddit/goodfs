package config

import (
	"common/logs"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
)

const (
	ConfFilePath = "../conf/admin-server.yaml"
)

type Config struct {
	Port         string `yaml:"port" env:"PORT" env-default:"80"`
	ResourcePath string `yaml:"resource-path""`
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
