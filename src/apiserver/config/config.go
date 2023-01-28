package config

import (
	"apiserver/internal/usecase/componet/auth"
	"common/cst"
	"common/datasize"
	"common/etcd"
	"common/logs"
	"common/performance"
	"common/registry"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	ConfFilePath = "../../conf/api-server.yaml"
)

type Config struct {
	Checksum       bool               `yaml:"checksum" env:"CHECKSUM" env-default:"false"`
	Port           string             `yaml:"port" env:"PORT" env-default:"8080"`
	SelectStrategy string             `yaml:"select-strategy" env:"SELECT_STRATEGY" env-default:"random"`
	LocateTimeout  time.Duration      `yaml:"locate-timeout" env:"LOCATE_TIMEOUT" env-default:"1s"`
	Log            logs.Config        `yaml:"log" env-prefix:"LOG"`
	Etcd           etcd.Config        `yaml:"etcd" env-prefix:"ETCD"`
	Object         ObjectConfig       `yaml:"object" env-prefix:"OBJECT"`
	Discovery      DiscoveryConfig    `yaml:"discovery" env-prefix:"DISCOVERY"`
	Registry       registry.Config    `yaml:"registry" env-prefix:"REGISTRY"`
	Auth           auth.Config        `yaml:"auth" env-prefix:"AUTH"`
	Performance    performance.Config `yaml:"performance" env-prefix:"PERFORMANCE"`
	TLS            TLSConfig          `yaml:"tls" env-prefix:"TLS"`
}

func (c *Config) initialize() {
	// aligned to 4KB
	if i := c.Object.ReedSolomon.BlockPerShard % cst.OS.PageSize; i > 0 {
		c.Object.ReedSolomon.BlockPerShard = c.Object.ReedSolomon.BlockPerShard - i + cst.OS.PageSize
	}
	if i := int(c.Object.Replication.BlockSize) % cst.OS.PageSize; i > 0 {
		c.Object.Replication.BlockSize = datasize.DataSize(int(c.Object.Replication.BlockSize) - i + cst.OS.PageSize)
	}
}

type DiscoveryConfig struct {
	DataServName string `yaml:"data-serv-name" env:"DATA_SERV_NAME" env-default:"objectserver"`
	MetaServName string `yaml:"meta-serv-name" env:"META_SERV_NAME" env-default:"metaserver"`
}

type ObjectConfig struct {
	ReedSolomon RsConfig          `yaml:"reed-solomon" env-prefix:"REED_SOLOMON"`
	Replication ReplicationConfig `yaml:"replication" env-prefix:"REPLICATION"`
}

type ReplicationConfig struct {
	CopiesCount       int               `yaml:"copies-count" env:"COPIES_COUNT" env-default:"3"`
	BlockSize         datasize.DataSize `yaml:"block-size" env:"BLOCK_SIZE" env-default:"32KB"` // Auto aligned to power of 4KB
	LossToleranceRate float32           `yaml:"loss-tolerance-rate" env:"LOSS_TOLERANCE_RATE" env-default:"0"`
	CopyAsync         bool              `yaml:"copy-async" env:"COPY_ASYNC" env-default:"false"`
}

func (rp *ReplicationConfig) AtLeastCopiesNum() int {
	return rp.CopiesCount - rp.ToleranceLossNum()
}

func (rp *ReplicationConfig) ToleranceLossNum() int {
	toleranceNum := rp.LossToleranceRate * float32(rp.CopiesCount)
	return int(toleranceNum)
}

type RsConfig struct {
	DataShards    int `yaml:"data-shards" env:"DATA_SHARDS" env-default:"4"`
	ParityShards  int `yaml:"parity-shards" env:"PARITY_SHARDS" env-default:"2"`
	BlockPerShard int `yaml:"block-per-shard" env:"BLOCK_PER_SHARD" env-default:"8192"` // Auto aligned to power of 4KB
}

func (r *RsConfig) AllShards() int {
	return r.DataShards + r.ParityShards
}

func (r *RsConfig) BlockSize() int {
	return r.BlockPerShard * r.DataShards
}

func (r *RsConfig) ShardSize(totalSize int64) int {
	dsNum := int64(r.DataShards)
	return int((totalSize + dsNum - 1) / dsNum)
}

type TLSConfig struct {
	Enabled        bool   `yaml:"enabled" env:"ENABLED"`
	ServerCertFile string `yaml:"server-cert-file" env:"SERVER_CERT_FILE"`
	ServerKeyFile  string `yaml:"server-key-file" env:"SERVER_KEY_FILE"`
}

func ReadConfig() Config {
	var conf Config
	if err := cleanenv.ReadConfig(ConfFilePath, &conf); err != nil {
		panic(err)
	}
	logs.Std().Infof("read config from %s", ConfFilePath)
	conf.initialize()
	return conf
}

func ReadConfigFrom(path string) Config {
	var conf Config
	if err := cleanenv.ReadConfig(path, &conf); err != nil {
		if os.IsNotExist(err) {
			return ReadConfig()
		}
		panic(err)
	}
	logs.Std().Infof("read config from %s", path)
	conf.initialize()
	return conf
}
