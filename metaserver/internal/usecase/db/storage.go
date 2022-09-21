package db

import (
	"common/graceful"
	"common/logs"
	"common/util"
	"metaserver/internal/usecase"
	"os"
	"sync/atomic"
	"time"

	bolt "go.etcd.io/bbolt"
)

var (
	dbLog = logs.New("storage")
)

type Storage struct {
	originalPath string
	current      atomic.Value
	rdOnly       atomic.Value
}

func NewStorage() *Storage {
	at := atomic.Value{}
	at.Store(false)
	return &Storage{rdOnly: at}
}

func (s *Storage) DB() *bolt.DB {
	return s.current.Load().(*bolt.DB)
}

func (s *Storage) View(fn usecase.TxFunc) error {
	return s.DB().View(fn)
}

func (s *Storage) Update(fn usecase.TxFunc) error {
	if s.rdOnly.Load().(bool) {
		return usecase.ErrReadOnly
	}
	return s.DB().Update(fn)
}

func (s *Storage) Stop() error {
	dbLog.Info("stop db...")
	curDB := s.DB()
	curPath := curDB.Path()
	if err := curDB.Close(); err != nil {
		return err
	}
	if curPath != s.originalPath {
		dbLog.Infof("db file has been replaced, rename '%s' to original '%s'", curPath, s.originalPath)
		return os.Rename(curPath, s.originalPath)
	}
	return nil
}

func (s *Storage) Open(path string) error {
	cur, err := bolt.Open(path, util.OS_ModeUser, &bolt.Options{
		Timeout:    12 * time.Second,
		NoGrowSync: false,
		FreelistType: bolt.FreelistMapType,
	})
	if err != nil {
		return err
	}
	s.originalPath = path
	s.current.Store(cur)
	return nil
}

func (s *Storage) Replace(replacePath string) (err error) {
	s.rdOnly.Store(true)
	defer s.rdOnly.Store(false)

	var newDB *bolt.DB
	if newDB, err = bolt.Open(replacePath, util.OS_ModeUser, &bolt.Options{
		Timeout:    12 * time.Second,
		NoGrowSync: false,
	}); err != nil {
		return err
	}
	// record current db
	old := s.DB()
	// change current db to new one
	s.current.Store(newDB)
	// close and remove old db file
	go func() {
		defer graceful.Recover()
		oldPath := old.Path()
		util.LogErrWithPre("close old db", old.Close())
		util.LogErrWithPre("remove old db file", os.RemoveAll(oldPath))
	}()
	return
}
