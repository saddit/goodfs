package auth

import (
	"gopkg.in/yaml.v3"
)

type Mode int8

const (
	ModeDisable Mode = 1 << iota
	ModePassword
	ModeCallback
)

func ParseMode(s string) Mode {
	switch s {
	case "password", "pwd":
		return ModePassword
	case "callback":
		return ModeCallback
	default:
		return ModeDisable
	}
}

func (m *Mode) UnmarshalYAML(node *yaml.Node) error {
	return m.SetValue(node.Value)
}

func (m *Mode) SetValue(s string) error {
	*m = ParseMode(s)
	return nil
}

type CallbackConfig struct {
	Url    string   `yaml:"url" env:"URL"`
	Params []string `yaml:"params" env:"PARAMS" env-seprator:","`
}

type Config struct {
	Mode     Mode           `yaml:"mode" env:"MODE" env-default:"disable"`
	Username string         `yaml:"username" env:"USERNAME"`
	Password string         `yaml:"password" env:"PASSWORD"`
	Callback CallbackConfig `yaml:"callback" env-prefx:"CALLBACK"`
}
