package db

import (
	"bytes"
	"common/logs"
	"common/util"
	"github.com/dgraph-io/badger/v3"
	"github.com/dgraph-io/badger/v3/options"
	"os"
	"runtime"
)

var (
	Sep    = []byte(",")
	SepEnc = []byte("%2C")
)

type PathCache struct {
	db *badger.DB
}

func NewPathCache(storePath string) (*PathCache, error) {
	db, err := badger.Open(
		badger.DefaultOptions(storePath).
			WithNumGoroutines(runtime.NumCPU()).
			WithCompression(options.ZSTD).
			WithLogger(logs.New("path-cache")),
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

func (pc *PathCache) encodeValue(val []byte) []byte {
	// FIXME: allocate N and copy N
	return bytes.ReplaceAll(val, Sep, SepEnc)
}

func (pc *PathCache) decodeValue(val []byte) []byte {
	// FIXME: allocate N and copy N
	return bytes.ReplaceAll(val, SepEnc, Sep)
}

func (pc *PathCache) Put(name, path string) error {
	return pc.db.Update(func(txn *badger.Txn) error {
		key := util.StrToBytes(name)
		value := pc.encodeValue(util.StrToBytes(path))
		origin, err := pc.get(txn, key)
		if err != nil && err != badger.ErrKeyNotFound {
			return err
		}
		// not use bytes.Join to avoid memory copy
		return txn.Set(key, append(append(origin, Sep...), value...))
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
		sp := bytes.Split(value, Sep)
		if len(sp) == 0 {
			return badger.ErrKeyNotFound
		}
		res = make([]string, 0, len(sp))
		for _, b := range sp {
			res = append(res, util.BytesToStr(pc.decodeValue(b)))
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
		if idx := bytes.LastIndexByte(value, Sep[0]); idx > 0 {
			res = util.BytesToStr(pc.decodeValue(value[idx+1:]))
			return nil
		}
		return badger.ErrKeyNotFound
	})
	if err == badger.ErrKeyNotFound {
		err = os.ErrNotExist
	}
	return res, err
}

func (pc *PathCache) Remove(name, path string) error {
	err := pc.db.Update(func(txn *badger.Txn) error {
		key := util.StrToBytes(name)
		val, err := pc.get(txn, key)
		if err != nil {
			return err
		}
		before, after, ok := bytes.Cut(val, pc.encodeValue(util.StrToBytes(path)))
		if !ok {
			return badger.ErrKeyNotFound
		}
		before = bytes.TrimSuffix(before, Sep)
		after = bytes.TrimPrefix(after, Sep)
		// not use bytes.Join to avoid memory allocating and copy
		return txn.Set(key, append(append(before, Sep...), after...))
	})
	if err == badger.ErrKeyNotFound {
		err = nil
	}
	return err
}

func (pc *PathCache) RemoveAll(name string) error {
	err := pc.db.Update(func(txn *badger.Txn) error {
		key := util.StrToBytes(name)
		return txn.Delete(key)
	})
	if err == badger.ErrKeyNotFound {
		err = nil
	}
	return err
}

func (pc *PathCache) IteratorAll(fn func(path string) error) error {
	return pc.db.View(func(txn *badger.Txn) error {
		itr := txn.NewIterator(badger.DefaultIteratorOptions)
		defer itr.Close()
		for itr.Valid() {
			itr.Next()
			if err := itr.Item().Value(func(val []byte) error {
				for _, b := range bytes.Split(val, Sep) {
					if err := fn(util.BytesToStr(pc.decodeValue(b))); err != nil {
						return err
					}
				}
				return nil
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func (pc *PathCache) Close() error {
	return pc.db.Close()
}
