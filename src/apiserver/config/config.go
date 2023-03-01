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
	Port           string             `yaml:"port" env:"PORT" env-default:"8080"`
	SelectStrategy string             `yaml:"select-strategy" env:"SELECT_STRATEGY" env-default:"random"`
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
	if i := c.Object.ReedSolomon.BlockSize() % cst.OS.NetPkgSize; i > 0 {
		newSize := c.Object.ReedSolomon.BlockSize() - i + cst.OS.NetPkgSize
		c.Object.ReedSolomon.BlockPerShard = newSize / c.Object.ReedSolomon.DataShards
		logs.Std().Warnf("aligned object.reedsolomon.block-per-shard to %d", c.Object.ReedSolomon.BlockPerShard)
	}
	if i := c.Object.Replication.BlockSize % datasize.DataSize(cst.OS.NetPkgSize); i > 0 {
		c.Object.Replication.BlockSize = c.Object.Replication.BlockSize - i + datasize.DataSize(cst.OS.NetPkgSize)
		logs.Std().Warnf("aligned object.replication.block-size to %d", c.Object.Replication.BlockSize)
	}
}

type DiscoveryConfig struct {
	DataServName string `yaml:"data-serv-name" env:"DATA_SERV_NAME" env-default:"objectserver"`
	MetaServName string `yaml:"meta-serv-name" env:"META_SERV_NAME" env-default:"metaserver"`
}

type ObjectConfig struct {
	Checksum        bool              `yaml:"checksum" env:"CHECKSUM"`
	DistinctSize    datasize.DataSize `yaml:"distinct-size" env:"DISTINCT_SIZE"`
	DistinctTimeout time.Duration     `yaml:"distinct-timeout" env:"DISTINCT_TIMEOUT" env-default:"200ms"`
	ReedSolomon     RsConfig          `yaml:"reed-solomon" env-prefix:"REED_SOLOMON"`
	Replication     ReplicationConfig `yaml:"replication" env-prefix:"REPLICATION"`
}

type ReplicationConfig struct {
	CopiesCount       int               `yaml:"copies-count" env:"COPIES_COUNT" env-default:"3"`
	BlockSize         datasize.DataSize `yaml:"block-size" env:"BLOCK_SIZE" env-default:"32KB"` // Auto aligned to power of 4KB
	LossToleranceRate float32           `yaml:"loss-tolerance-rate" env:"LOSS_TOLERANCE_RATE" env-default:"0"`
	CopyAsync         bool              `yaml:"copy-async" env:"COPY_ASYNC" env-default:"true"`
}

func (rp *ReplicationConfig) AtLeastCopiesNum() int {
	return rp.CopiesCount - rp.ToleranceLossNum()
}

func (rp *ReplicationConfig) ToleranceLossNum() int {
	toleranceNum := rp.LossToleranceRate * float32(rp.CopiesCount)
	return int(toleranceNum)
}

type RsConfig struct {
	DataShards    int  `yaml:"data-shards" env:"DATA_SHARDS" env-default:"4"`             // DataShards shards number of data part
	ParityShards  int  `yaml:"parity-shards" env:"PARITY_SHARDS" env-default:"2"`         // ParityShards shards number of parity part
	BlockPerShard int  `yaml:"block-per-shard" env:"BLOCK_PER_SHARD" env-default:"16384"` // BlockPerShard auto increase to make BlockSize is multiple of 16KB
	RewriteAsync  bool `yaml:"rewrite-async" env:"REWRITE_ASYNC" env-default:"true"`      // RewriteAsync store lost shard asynchronously
}

func (r *RsConfig) AllShards() int {
	return r.DataShards + r.ParityShards
}

func (r *RsConfig) BlockSize() int {
	return r.BlockPerShard * r.DataShards
}

func (r *RsConfig) FullSize() int {
	return r.BlockPerShard * r.AllShards()
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
