package registry

import "time"

type Config struct {
	ServerID   string        `yaml:"server-id" env:"SERVER_ID" env-required:"true"`
	Group      string        `yaml:"group" env:"GROUP" env-default:"goodfs"`
	Name       string        `yaml:"name" env:"NAME" env-required:"true"`
	Interval   time.Duration `yaml:"interval" env:"INTERVAL" env-default:"5s"`
	Timeout    time.Duration `yaml:"timeout" env:"TIMEOUT" env-default:"3s"`
	Services   []string      `yaml:"services,omitempty" env:"SERVICES" env-separator:","`
	ServerPort string        `yaml:"-" env:"-"`
}
