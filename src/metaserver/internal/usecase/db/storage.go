package db

import (
	"common/cst"
	"common/graceful"
	"common/logs"
	"common/util"
	"io/fs"
	"metaserver/internal/usecase"
	"os"
	"path/filepath"
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
	if logs.IsDebug() {
		start := time.Now()
		defer func() { dbLog.Debugf("read-only tx spent %d ms", time.Since(start).Milliseconds()) }()
	}
	return s.DB().View(fn)
}

func (s *Storage) Update(fn usecase.TxFunc) error {
	if logs.IsDebug() {
		start := time.Now()
		defer func() { dbLog.Debugf("read-write tx spent %d ms", time.Since(start).Milliseconds()) }()
	}
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
		dbLog.Infof("db file has been replaced, rename original '%s' to '%s'", s.originalPath, curPath)
		return os.Rename(curPath, s.originalPath)
	}
	return nil
}

func (s *Storage) FileInfo() (fs.FileInfo, error) {
	return os.Stat(s.DB().Path())
}

func (s *Storage) checkPath(path string) error {
	dir := filepath.Dir(path)
	_, err := os.Stat(dir)
	if err != nil && os.IsNotExist(err) {
		return os.MkdirAll(dir, cst.OS.ModeUser)
	}
	return err
}

func (s *Storage) Open(path string) error {
	if err := s.checkPath(path); err != nil {
		return err
	}
	cur, err := bolt.Open(path, cst.OS.ModeUser, &bolt.Options{
		Timeout:      12 * time.Second,
		NoGrowSync:   false,
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
	if newDB, err = bolt.Open(replacePath, cst.OS.ModeUser, &bolt.Options{
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
