package db

import (
	"bytes"
	"common/logs"
	"common/util"
	"common/util/slices"
	"github.com/dgraph-io/badger/v3"
	"github.com/dgraph-io/badger/v3/options"
	"os"
	"runtime"
)

type PathCache struct {
	db *badger.DB
}

func NewPathCache(storePath string) (*PathCache, error) {
	db, err := badger.Open(
		badger.DefaultOptions(storePath).
			WithNumGoroutines(runtime.NumCPU()).
			WithCompression(options.ZSTD).
			WithLogger(logs.Std()),
	)
	if err != nil {
		return nil, err
	}
	return &PathCache{
		db: db,
	}, nil
}

func (pc *PathCache) get(txn *badger.Txn, key []byte) ([]byte, error) {
	var res []byte
	item, err := txn.Get(key)
	if err == nil {
		if err = item.Value(func(val []byte) error {
			res = val
			return nil
		}); err != nil {
			return nil, err
		}
	}
	return res, err
}

func (pc *PathCache) Put(name, path string) error {
	return pc.db.Update(func(txn *badger.Txn) error {
		key := util.StrToBytes(name)
		value := util.StrToBytes(path)
		origin, err := pc.get(txn, key)
		if err != nil && err != badger.ErrKeyNotFound {
			return err
		}
		return txn.Set(key, bytes.Join([][]byte{origin, value}, []byte{','}))
	})
}

// Get if not found, return os.ErrNotExist
func (pc *PathCache) Get(name string) ([]string, error) {
	var res []string
	err := pc.db.View(func(txn *badger.Txn) error {
		key := util.StrToBytes(name)
		value, err := pc.get(txn, key)
		if err != nil {
			return err
		}
		sp := bytes.Split(value, []byte{','})
		if len(sp) == 0 {
			return badger.ErrKeyNotFound
		}
		res = make([]string, 0, len(sp))
		for _, b := range sp {
			res = append(res, util.BytesToStr(b))
		}
		return nil
	})
	if err == badger.ErrKeyNotFound {
		err = os.ErrNotExist
	}
	return res, err
}

// GetLast if not found, return os.ErrNotExist
func (pc *PathCache) GetLast(name string) (string, error) {
	var res string
	err := pc.db.View(func(txn *badger.Txn) error {
		key := util.StrToBytes(name)
		value, err := pc.get(txn, key)
		if err != nil {
			return err
		}
		sp := bytes.Split(value, []byte{','})
		if len(sp) == 0 {
			return badger.ErrKeyNotFound
		}
		res = util.BytesToStr(slices.Last(sp))
		return nil
	})
	if err == badger.ErrKeyNotFound {
		err = os.ErrNotExist
	}
	return res, err
}

func (pc *PathCache) Close() error {
	return pc.db.Close()
}
