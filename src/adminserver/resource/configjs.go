package resource

// Special content for embed file 'config.js'

import (
	"adminserver/internal/usecase/pool"
	"common/util"
	"fmt"
	"io/fs"
	"net/http"
)

type configJSInfo struct {
	fs.FileInfo
	size int64
}

func (ci configJSInfo) Size() int64 {
	return ci.size
}

type configJS struct {
	http.File
	content []byte
}

func newConfigJS(file http.File) configJS {
	content := fmt.Sprintf(`window.baseUrl = "%s://%s/api"`, "http", util.ServerAddress(pool.Config.Port))
	return configJS{file, util.StrToBytes(content)}
}

func (cj configJS) Stat() (fs.FileInfo, error) {
	info, err := cj.File.Stat()
	return configJSInfo{FileInfo: info, size: int64(len(cj.content))}, err
}

func (cj configJS) Read(p []byte) (n int, err error) {
	n = copy(p, cj.content)
	return
}
