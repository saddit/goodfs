package entity

type VerMode int32

const (
	//VerModeFirst 查询第一个版本
	VerModeFirst VerMode = 1
	//VerModeLast 只查询最后一个版本
	VerModeLast VerMode = 0
	// VerModeNot 不查询任何版本
	VerModeNot VerMode = -1
)

type ObjectStrategy int8

const (
	ECReedSolomon ObjectStrategy = 1 << iota
	MultiReplication
)

type Metadata struct {
	Name       string     `json:"name"`
	Bucket     string     `json:"bucket"`
	CreateTime int64      `json:"createTime"`
	UpdateTime int64      `json:"updateTime"`
	Versions   []*Version `json:"versions"`
}

func (m *Metadata) LastVersion() *Version {
	if len(m.Versions) > 0 {
		return m.Versions[len(m.Versions)-1]
	}
	return nil
}

type Version struct {
	Hash          string         `json:"hash"`
	StoreStrategy ObjectStrategy `json:"storeStrategy"`
	Sequence      int32          `json:"sequence"`
	Size          int64          `json:"size"`
	Ts            int64          `json:"ts"`
	DataShards    int            `json:"dataShards"`
	ParityShards  int            `json:"parityShards"`
	ShardSize     int            `json:"shardSize"`
	Locate        []string       `json:"locate"`
}

type Bucket struct {
	Versioning     bool           `json:"versioning"`     // Versioning marks bucket can store multi versions of object. if true, VersionRemains will be used
	Readonly       bool           `json:"readonly"`       // Readonly marks objects in bucket only allowed to read
	StoreStrategy  ObjectStrategy `json:"storeStrategy"`  // StoreStrategy if not zero, it will apply to ever objects under this bucket
	DataShards     int            `json:"dataShards"`     // DataShards used when StoreStrategy is not zero
	ParityShards   int            `json:"parityShards"`   // ParityShards used when StoreStrategy is not zero
	VersionRemains int            `json:"versionRemains"` // VersionRemains is maximum number of remained versions
	CreateTime     int64          `json:"createTime"`     // CreateTime is bucket created time
	UpdateTime     int64          `json:"updateTime"`     // UpdateTime is last updating time
	Name           string         `json:"name"`           // Name is the bucket's name
	Policies       []string       `json:"policies"`       // Policies is the iam polices for this bucket (No support yet)
}
