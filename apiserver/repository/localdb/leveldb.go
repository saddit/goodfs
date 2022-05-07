package localdb

import (
	"encoding/json"
	"goodfs/apiserver/global"
)

func Insert(key string, value any) error {
	bt, e := json.Marshal(value)
	if e != nil {
		return e
	}
	e = global.LocalDB.Put([]byte(key), bt, nil)
	return e
}

func Delete(key string) error {
	return global.LocalDB.Delete([]byte(key), nil)
}

func Get[T any](key string) *T {
	value, e := global.LocalDB.Get([]byte(key), nil)
	if e != nil || value == nil {
		return nil
	}
	var v T
	if e = json.Unmarshal(value, &v); e != nil {
		return nil
	}
	return &v
}
