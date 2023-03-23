package registry

import (
	"common/util"
	"fmt"
	"net"
	"strings"
	"time"
)

type Config struct {
	ServerIP   string        `yaml:"server-ip" env:"SERVER_IP"`
	ServerID   string        `yaml:"server-id" env:"SERVER_ID"`
	Group      string        `yaml:"group" env:"GROUP" env-default:"goodfs"`
	Name       string        `yaml:"name" env:"NAME" env-required:"true"`
	Interval   time.Duration `yaml:"interval" env:"INTERVAL" env-default:"5s"`
	Timeout    time.Duration `yaml:"timeout" env:"TIMEOUT" env-default:"3s"`
	Services   []string      `yaml:"services,omitempty" env:"SERVICES" env-separator:","`
	ServerPort string        `yaml:"-" env:"-"`
}

func (c *Config) RegisterAddr() (string, bool) {
	if c.ServerPort == "" {
		panic("registry required ServerPort")
	}
	if c.ServerIP == "" {
		return util.ServerAddress(c.ServerPort), false
	}
	return net.JoinHostPort(c.ServerIP, c.ServerPort), true
}

func (c *Config) RegisterKey() string {
	return fmt.Sprint(c.Group, "/", c.Name, "/", c.SID())
}

func (c *Config) SID() string {
	if c.ServerID == "" {
		addr, _ := c.RegisterAddr()
		addr = strings.ReplaceAll(addr, ".", "")
		addr = strings.ReplaceAll(addr, ":", "_")
		c.ServerID = fmt.Sprint(c.Name, "-", addr)
	}
	return c.ServerID
}
