package resource

import (
	"common/logs"
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-contrib/static"
)

//go:embed web
var embedFs embed.FS

type embedFileSystem struct {
	http.FileSystem
	indexes bool
}

func (e embedFileSystem) Exists(prefix string, path string) bool {
	f, err := e.Open(path)
	if err != nil {
		logs.Std().Errorf("bad static path: %s, %s", path, err)
		return false
	}

	// check if indexing is allowed
	s, _ := f.Stat()
	if s.IsDir() && !e.indexes {
		return false
	}

	return true
}

func FileSystem() static.ServeFileSystem {
	fsys, err := fs.Sub(embedFs, "web")
	if err != nil {
		panic(err)
	}
	return embedFileSystem{
		FileSystem: http.FS(fsys),
		indexes:    false,
	}
}
