package etcd

type Config struct {
	Endpoint []string `yaml:"endpoint" env:"ENDPOINT" env-required:"true" env-separator:","`
	Username string   `yaml:"username" env:"USERNAME" env-required:"true"`
	Password string   `yaml:"password" env:"PASSWORD" env-required:"true"`
}
