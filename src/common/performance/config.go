package performance

import "time"

type StoreType string

const (
	None   StoreType = "none"
	Local  StoreType = "local"
	Remote StoreType = "remote"
)

type Config struct {
	Enable           bool          `yaml:"enable" env:"ENABLE" env-default:"false"`                         // Enable should enable performance info collection. default is false
	Store            StoreType     `yaml:"store" env:"STORE" env-default:"none"`                            // Store specify a storage type. default is none (not store anything)
	FlushInterval    time.Duration `yaml:"flush-interval" env:"FLUSH_INTERVAL" env-default:"5m"`            // FlushInterval the interval to flush in-memory records to Store. it will be reset by FlushWhenReached. default is 5 minute.
	MaxInMemory      int           `yaml:"max-in-memory" env:"MAX_IN_MEMORY" env-default:"1000"`            // MaxInMemory is maximum records allowed to stay in memory, default is 1000.
	FlushWhenReached bool          `yaml:"flush-when-reached" env:"FLUSH_WHEN_REACHED" env-default:"false"` // FlushWhenReached flush immediately when in-memory size reach limitation if true. otherwise the first in will first out. default is false.
}
