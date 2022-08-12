package db

import "go.etcd.io/bbolt"

type Storage struct {
	*bbolt.DB
}

//TODO wrapper db