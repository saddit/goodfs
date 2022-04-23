package config

import (
	"os"
	"time"

	yaml "gopkg.in/yaml.v3"
)

const (
	ConfFilePath = "conf/api-server.yaml"
)

type Config struct {
	Port           string        `yaml:"port"`
	AmqpAddress    string        `yaml:"amqp_address"`
	MongoAddress   string        `yaml:"mongo_address"`
	LogDir         string        `yaml:"log_dir"`
	DetectInterval time.Duration `yaml:"detect_interval"`
	SuspendTimeout time.Duration `yaml:"suspend_timeout"`
	DeadTimeout    time.Duration `yaml:"dead_timeout"`
	SelectStrategy string        `yaml:"select_strategy"`
	MachineCode    string        `yaml:"machine_code"`
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
