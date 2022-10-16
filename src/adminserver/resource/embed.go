package resource

import (
	"embed"
	"github.com/gin-contrib/static"
	"io/fs"
	"net/http"
)

//go:embed web
var embedFs embed.FS

type embedFileSystem struct {
	http.FileSystem
}

func (e embedFileSystem) Exists(prefix string, path string) bool {
	_, err := e.Open(path)
	if err != nil {
		return false
	}
	return true
}

func FileSystem(targetPath string) static.ServeFileSystem {
	fsys, err := fs.Sub(embedFs, targetPath)
	if err != nil {
		panic(err)
	}
	return embedFileSystem{
		FileSystem: http.FS(fsys),
	}
}
