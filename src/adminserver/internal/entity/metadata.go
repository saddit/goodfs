package entity

type Metadata struct {
	Name       string `json:"name"`
	CreateTime int64  `json:"createTime"`
	UpdateTime int64  `json:"updateTime"`
}

type Bucket struct {
	Versioning     bool     `json:"versioning"`     // Versioning marks bucket can store multi versions of object. if true, VersionRemains will be used
	Readonly       bool     `json:"readonly"`       // Readonly marks objects in bucket only allowed to read
	DataShards     int      `json:"dataShards"`     // DataShards used when StoreStrategy is not zero
	ParityShards   int      `json:"parityShards"`   // ParityShards used when StoreStrategy is not zero
	VersionRemains int      `json:"versionRemains"` // VersionRemains is maximum number of remained versions
	StoreStrategy  int8     `json:"storeStrategy"`  // StoreStrategy if not zero, it will apply to ever objects under this bucket
	CreateTime     int64    `json:"createTime"`     // CreateTime is bucket created time
	UpdateTime     int64    `json:"updateTime"`     // UpdateTime is last updating time
	Name           string   `json:"name"`           // Name is the bucket's name
	Policies       []string `json:"policies"`       // Policies is the iam polices for this bucket (No support yet)
}
