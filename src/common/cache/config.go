package cache

import (
	"common/datasize"
	"time"
)

type Config struct {
	TTL           time.Duration     `yaml:"ttl" env:"TTL" env-default:"20m"`
	CleanInterval time.Duration     `yaml:"clean-interval" env:"CLEAN_INTERVAL" env-default:"10m"`
	MaxSize       datasize.DataSize `yaml:"max-size" env:"MAX_SIZE" env-default:"128MB"`
}
