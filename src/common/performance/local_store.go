package performance

import (
	"bytes"
	"common/logs"
	"common/util"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	bolt "go.etcd.io/bbolt"
)

var (
	actionBucketRoot = []byte("action_bucket_root")
	kindBucketRoot   = []byte("kind_bucket_root")
	errNotExist      = errors.New("bucket or key not exist")
)

// LocalStore simply implements Store by using boltdb.
// one cost-time will be copied twice.
type LocalStore struct {
	db   *bolt.DB
	mux  sync.Locker
	path string
}

func NewLocalStore(path string) Store {
	underlying := &LocalStore{mux: &sync.Mutex{}, path: path}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil && !os.IsExist(err) {
		logs.Std().Errorf("open boltdb err: %s", err)
		return nil
	}
	if err := underlying.openDatabase(path); err != nil {
		logs.Std().Errorf("open boltdb err: %s", err)
		return nil
	}
	return AvgSumStore(underlying)
}

func (ls *LocalStore) getBucket(tx *bolt.Tx, rootName []byte, name string) (b *bolt.Bucket, e error) {
	defer func() {
		if e == nil {
			b.FillPercent = 0.9
		}
	}()
	if tx.Writable() {
		root, err := tx.CreateBucketIfNotExists(rootName)
		if err != nil {
			return nil, err
		}
		if name == "" {
			return root, nil
		}
		return root.CreateBucketIfNotExists(util.StrToBytes(name))
	}
	root := tx.Bucket(rootName)
	if root == nil {
		return nil, errNotExist
	}
	if name == "" {
		return root, nil
	}
	bucket := root.Bucket(util.StrToBytes(name))
	if bucket == nil {
		return nil, errNotExist
	}
	return bucket, nil
}

func (ls *LocalStore) Put(data []*Perform) error {
	if err := ls.checkDatabase(); err != nil {
		return err
	}
	return ls.db.Batch(func(tx *bolt.Tx) error {
		for _, item := range data {
			// get bucket
			actBuk, err := ls.getBucket(tx, actionBucketRoot, item.Action)
			if err != nil {
				return err
			}
			kndBuk, err := ls.getBucket(tx, kindBucketRoot, item.KindOf)
			if err != nil {
				return err
			}
			// parse value
			value := util.IntToBytes(uint64(item.Cost))
			// save to action-bucket
			i, _ := actBuk.NextSequence()
			actKey := util.StrToBytes(fmt.Sprint(item.KindOf, ".", i))
			err = actBuk.Put(actKey, value)
			if err != nil {
				return err
			}
			// save to kind-bucket
			i, _ = kndBuk.NextSequence()
			kndKey := util.StrToBytes(fmt.Sprint(item.Action, ".", i))
			err = kndBuk.Put(kndKey, value)
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func (ls *LocalStore) keyPrefix(key string) string {
	sp := strings.Split(key, ".")
	if len(sp) == 2 {
		return sp[0]
	}
	return ""
}

func (ls *LocalStore) Get(kind, action string) ([]*Perform, error) {
	if err := ls.checkDatabase(); err != nil {
		return nil, err
	}
	var res []*Perform
	err := ls.db.View(func(tx *bolt.Tx) error {
		if kind != "" {
			kndBuk, err := ls.getBucket(tx, kindBucketRoot, kind)
			if err != nil {
				return err
			}
			var k, v []byte
			cur := kndBuk.Cursor()
			if action != "" {
				k, v = cur.Seek(util.StrToBytes(action))
			} else {
				k, v = cur.First()
			}
			for k != nil {
				act := ls.keyPrefix(util.BytesToStr(k))
				if action != "" && act != action {
					break
				}
				res = append(res, &Perform{
					KindOf: kind,
					Action: act,
					Cost:   time.Duration(util.BytesToInt(v)),
				})
				k, v = cur.Next()
			}
			return nil
		} else {
			// if action is empty returns action root bucket
			actBuk, err := ls.getBucket(tx, actionBucketRoot, action)
			if err != nil {
				return err
			}
			return actBuk.ForEach(func(k, v []byte) error {
				if v == nil {
					// if k is a nest bucket (when action is emtpy, action root bucket has nest buckets)
					actBuk.Bucket(k).ForEach(func(k, v []byte) error {
						res = append(res, &Perform{
							KindOf: ls.keyPrefix(util.BytesToStr(k)),
							Action: util.BytesToStr(k),
							Cost:   time.Duration(util.BytesToInt(v)),
						})
						return nil
					})
					return nil
				}
				res = append(res, &Perform{
					KindOf: ls.keyPrefix(util.BytesToStr(k)),
					Action: action,
					Cost:   time.Duration(util.BytesToInt(v)),
				})
				return nil
			})
		}
	})
	return res, err
}

func (ls *LocalStore) delete(tx *bolt.Tx, rootName []byte, nestName string, prefix []byte) error {
	if nestName != "" {
		nest, err := ls.getBucket(tx, rootName, nestName)
		if err == errNotExist {
			return nil
		}
		if err != nil {
			return err
		}
		cur := nest.Cursor()
		bk, _ := cur.Seek(prefix)
		for bk != nil {
			if !bytes.HasPrefix(bk, prefix) {
				break
			}
			if err = cur.Delete(); err != nil {
				return err
			}
			bk, _ = cur.Next()
		}
		return nil
	}
	// get root bucket
	root, err := ls.getBucket(tx, rootName, "")
	if err == errNotExist {
		return nil
	}
	if err != nil {
		return err
	}
	// remove from root
	return root.ForEach(func(k, v []byte) error {
		if v != nil {
			return nil
		}
		cur := root.Bucket(k).Cursor()
		bk, _ := cur.Seek(prefix)
		for bk != nil {
			if !bytes.HasPrefix(bk, prefix) {
				break
			}
			if err = cur.Delete(); err != nil {
				return err
			}
			bk, _ = cur.Next()
		}
		return nil
	})
}

func (ls *LocalStore) checkDatabase() error {
	ls.mux.Lock()
	defer ls.mux.Unlock()
	if ls.db == nil {
		return ls.openDatabase(ls.path)
	}
	return nil
}

func (ls *LocalStore) openDatabase(path string) error {
	opt := bolt.DefaultOptions
	opt.FreelistType = bolt.FreelistMapType
	db, err := bolt.Open(path, 0700, opt)
	if err != nil {
		return err
	}
	ls.db = db
	return nil
}

func (ls *LocalStore) Clear(kind, action string) error {
	if err := ls.checkDatabase(); err != nil {
		return err
	}
	if kind != "" && action == "" {
		return ls.db.Update(func(tx *bolt.Tx) error {
			// get kind root bucket
			kindRoot, err := ls.getBucket(tx, kindBucketRoot, "")
			if err == errNotExist {
				return nil
			}
			if err != nil {
				return err
			}
			// delete kind bucket directly
			if err = kindRoot.Delete(util.StrToBytes(kind)); err != nil {
				return err
			}
			// delete from action buckets
			return ls.delete(tx, actionBucketRoot, action, util.StrToBytes(kind))
		})
	} else if kind == "" && action != "" {
		return ls.db.Update(func(tx *bolt.Tx) error {
			// get action root bucket
			actionRoot, err := ls.getBucket(tx, actionBucketRoot, "")
			if err == errNotExist {
				return nil
			}
			if err != nil {
				return err
			}
			// delete action bucket directly
			if err = actionRoot.Delete(util.StrToBytes(action)); err != nil {
				return err
			}
			// delete from kind buckets
			return ls.delete(tx, kindBucketRoot, kind, util.StrToBytes(action))
		})
	} else if action != "" && kind != "" {
		kindBt := util.StrToBytes(kind)
		actionBt := util.StrToBytes(action)
		return ls.db.Update(func(tx *bolt.Tx) error {
			err := ls.delete(tx, actionBucketRoot, action, kindBt)
			if err != nil {
				return err
			}
			return ls.delete(tx, kindBucketRoot, kind, actionBt)
		})
	} else {
		// close db
		if err := ls.db.Close(); err != nil {
			return err
		}
		// remove db file
		path := ls.db.Path()
		if err := os.Remove(path); err != nil {
			return err
		}
		// re-open db
		return ls.openDatabase(path)
	}
}

func (ls *LocalStore) Size(kind, action string) (int64, error) {
	if err := ls.checkDatabase(); err != nil {
		return 0, err
	}
	var total int64
	err := ls.db.View(func(tx *bolt.Tx) error {
		if action != "" && kind == "" {
			buk, err := ls.getBucket(tx, actionBucketRoot, action)
			if err == errNotExist {
				return nil
			}
			if err != nil {
				return err
			}
			total = int64(buk.Stats().KeyN)
			return nil
		}
		if action == "" && kind != "" {
			buk, err := ls.getBucket(tx, kindBucketRoot, kind)
			if err == errNotExist {
				return nil
			}
			if err != nil {
				return err
			}
			total = int64(buk.Stats().KeyN)
			return nil
		}
		if action == "" && kind == "" {
			actRoot, err := ls.getBucket(tx, actionBucketRoot, "")
			if err == errNotExist {
				return nil
			}
			if err != nil {
				return err
			}
			return actRoot.ForEach(func(k, v []byte) error {
				if v != nil {
					return nil
				}
				total += int64(actRoot.Bucket(k).Stats().KeyN)
				return nil
			})
		}
		if action != "" && kind != "" {
			data, err := ls.Get(kind, action)
			if err == errNotExist {
				return nil
			}
			if err != nil {
				return err
			}
			total = int64(len(data))
			return nil
		}
		return nil
	})
	return total, err
}

func (ls *LocalStore) Average(string, string) ([]*Perform, error) {
	panic("not impl Average")
}

func (ls *LocalStore) Sum(string, string) ([]*Perform, error) {
	panic("not impl Sum")
}
