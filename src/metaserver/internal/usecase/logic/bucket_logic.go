package logic

import (
	"bytes"
	"common/util"
	"errors"
	bolt "go.etcd.io/bbolt"
	"metaserver/internal/entity"
	"metaserver/internal/usecase"
)

const (
	BucketBucketRoot = "go.dfs.bucket.root"
)

type BucketCrud struct {
}

func NewBucketCrud() *BucketCrud {
	return &BucketCrud{}
}

func (b *BucketCrud) getBucketBucket(tx *bolt.Tx) (*bolt.Bucket, error) {
	if tx.Writable() {
		return tx.CreateBucketIfNotExists(util.StrToBytes(BucketBucketRoot))
	}
	bk := tx.Bucket(util.StrToBytes(BucketBucketRoot))
	if bk == nil {
		return nil, usecase.ErrNotFound
	}
	bk.FillPercent = 0.9
	return bk, nil
}

func (b *BucketCrud) Get(name string, data *entity.Bucket) usecase.TxFunc {
	return func(tx *bolt.Tx) error {
		root, err := b.getBucketBucket(tx)
		if err != nil {
			return err
		}
		v := root.Get(util.StrToBytes(name))
		if v == nil {
			return usecase.ErrNotFound
		}
		return util.DecodeMsgp(data, v)
	}
}

func (b *BucketCrud) GetBytes(name string, bt *[]byte) usecase.TxFunc {
	return func(tx *bolt.Tx) error {
		root, err := b.getBucketBucket(tx)
		if err != nil {
			return err
		}
		*bt = root.Get(util.StrToBytes(name))
		if *bt == nil {
			return usecase.ErrNotFound
		}
		return nil
	}
}

func (b *BucketCrud) Create(data *entity.Bucket) usecase.TxFunc {
	return func(tx *bolt.Tx) error {
		if data.Name == "" {
			return errors.New("empty primary key 'Name'")
		}
		key := util.StrToBytes(data.Name)
		root, err := b.getBucketBucket(tx)
		if err != nil {
			return err
		}
		if root.Get(key) != nil {
			return usecase.ErrExists
		}
		v, err := util.EncodeMsgp(data)
		if err != nil {
			return err
		}
		return root.Put(key, v)
	}
}

func (b *BucketCrud) Delete(name string) usecase.TxFunc {
	return func(tx *bolt.Tx) error {
		root, err := b.getBucketBucket(tx)
		if err != nil {
			return err
		}
		return root.Delete(util.StrToBytes(name))
	}
}

func (b *BucketCrud) Update(data *entity.Bucket) usecase.TxFunc {
	return func(tx *bolt.Tx) error {
		if data.Name == "" {
			return errors.New("empty primary key 'Name'")
		}
		var origin entity.Bucket
		if err := b.Get(data.Name, &origin)(tx); err != nil {
			return err
		}
		if data.UpdateTime <= origin.UpdateTime {
			return usecase.ErrOldData
		}
		root, _ := b.getBucketBucket(tx)
		// update content
		data.CreateTime = origin.CreateTime
		v, err := util.EncodeMsgp(data)
		if err != nil {
			return err
		}
		return root.Put(util.StrToBytes(data.Name), v)
	}
}

func (b *BucketCrud) List(prefix string, limit int, res *[]*entity.Bucket, total *int) usecase.TxFunc {
	prefixBt := util.StrToBytes(prefix)
	return func(tx *bolt.Tx) error {
		root, err := b.getBucketBucket(tx)
		if err != nil {
			return err
		}
		cur := root.Cursor()
		var k, v []byte
		if prefix != "" {
			k, v = cur.Seek(prefixBt)
			defer func() { *total = len(*res) }()
		} else {
			*total = root.Stats().KeyN
			k, v = cur.First()
		}
		for k != nil && len(*res) < limit {
			if v == nil {
				continue
			}
			if prefix != "" && !bytes.HasPrefix(k, prefixBt) {
				break
			}
			var i entity.Bucket
			if err = util.DecodeMsgp(&i, v); err != nil {
				return err
			}
			*res = append(*res, &i)
			k, v = cur.Next()
		}
		return nil
	}
}

func (b *BucketCrud) Foreach(fn func(k, v []byte) error) usecase.TxFunc {
	return func(tx *bolt.Tx) error {
		root, err := b.getBucketBucket(tx)
		if err != nil {
			return err
		}
		return root.ForEach(func(k, v []byte) error {
			if v == nil {
				return nil
			}
			if err = fn(k, v); err != nil {
				return err
			}
			return nil
		})
	}
}
