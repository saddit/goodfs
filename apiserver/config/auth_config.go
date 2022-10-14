package config

type CallbackConfig struct {
	Enable bool     `yaml:"enable" env:"ENABLE"`
	Url    string   `yaml:"url" env:"URL"`
	Params []string `yaml:"params" env:"PARAMS" env-separator:","`
}

type PasswordConfig struct {
	Enable   bool   `yaml:"enable" env:"ENABLE"`
	Username string `yaml:"username" env:"USERNAME"`
	Password string `yaml:"password" env:"PASSWORD"`
}

type AuthConfig struct {
	Enable   bool           `yaml:"enable" env:"ENABLE"`
	Callback CallbackConfig `yaml:"callback" env-prefix:"CALLBACK"`
	Password PasswordConfig `yaml:"password" env-prefix:"PASSWORD"`
}
