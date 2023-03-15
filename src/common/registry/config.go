package registry

import (
	"fmt"
	"net"
	"time"
)

type Config struct {
	ServerIP   string        `yaml:"server-ip" env:"SERVER_IP"`
	ServerID   string        `yaml:"server-id" env:"SERVER_ID" env-required:"true"`
	Group      string        `yaml:"group" env:"GROUP" env-default:"goodfs"`
	Name       string        `yaml:"name" env:"NAME" env-required:"true"`
	Interval   time.Duration `yaml:"interval" env:"INTERVAL" env-default:"5s"`
	Timeout    time.Duration `yaml:"timeout" env:"TIMEOUT" env-default:"3s"`
	Services   []string      `yaml:"services,omitempty" env:"SERVICES" env-separator:","`
	ServerPort string        `yaml:"-" env:"-"`
}

func (c *Config) RegisterAddr() (string, bool) {
	if c.ServerIP == "" {
		return "", false
	}
	return net.JoinHostPort(c.ServerIP, c.ServerPort), true
}

func (c *Config) RegisterKey() string {
	return fmt.Sprint(c.Group, "/", c.Name, "/", c.ServerID)
}
