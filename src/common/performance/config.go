package performance

import "time"

type StoreType string

const (
	None   StoreType = "none"
	Local  StoreType = "local"
	Remote StoreType = "remote"
)

type Config struct {
	Enable          bool          `yaml:"enable" env:"ENABLE" env-default:"false"`              // Enable should enable performance info collection. default is false
	Store           StoreType     `yaml:"store" env:"STORE" env-default:"none"`                 // Store specify a storage type. default is none (not store anything)
	SaveInterval    time.Duration `yaml:"save-interval" env:"SAVE_INTERVAL" env-default:"5m"`   // SaveInterval the interval to put in-memory records to store. it will be reset by other put behavior. default is 5 minute
	MaxInMemory     int           `yaml:"max-in-memory" env:"MAX_IN_MEMORY" env-default:"1000"` // MaxInMemory is maximum records allowed to stay in memory, default is 1000.
	SaveWhenReached bool          `yaml:"save-when-reached" env-default:"false"`                // SaveWhenReached saving immediately when mem records reach max size if true. otherwise the first in will first out. default is false.
}
