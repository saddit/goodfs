package logs

type Config struct {
	Level    Level  `yaml:"level" env:"LEVEL" env-default:"info"`
	StoreDir string `yaml:"store_dir" env:"STORE_DIR"`
}
