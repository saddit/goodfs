package service

import (
	"apiserver/config"
	"apiserver/internal/usecase/webapi"
	"common/util"
	"encoding/base64"
	"fmt"
)

type resumeToken struct {
	Name    string           `json:"name"`
	Hash    string           `json:"hash"`
	Size    int64            `json:"size"`
	Servers []string         `json:"servers"`
	Ids     []string         `json:"ids"`
	Config  *config.RsConfig `json:"config"`
}

// RSResumablePutStream 断点续传
type RSResumablePutStream struct {
	*RSPutStream
	*resumeToken
}

// NewRSResumablePutStreamFromToken 恢复一个断点续传
func NewRSResumablePutStreamFromToken(token string) (*RSResumablePutStream, error) {
	bt, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, err
	}
	var tk resumeToken
	if ok := util.GobDecode(bt, &tk); ok {
		return &RSResumablePutStream{newExistedRSPutStream(tk.Servers, tk.Ids, tk.Hash, tk.Config), &tk}, nil
	}
	return nil, fmt.Errorf("invalid token")
}

// NewRSResumablePutStream 开启新的断点续传
func NewRSResumablePutStream(opt *StreamOption, rsCfg *config.RsConfig) (*RSResumablePutStream, error) {
	putStream, e := NewRSPutStream(opt, rsCfg)
	if e != nil {
		return nil, e
	}
	ids := make([]string, rsCfg.AllShards())
	for i := range ids {
		ids[i] = putStream.writers[i].(*PutStream).tmpId
	}
	token := &resumeToken{
		Name:    opt.Name,
		Hash:    opt.Hash,
		Servers: opt.Locates,
		Size:    opt.Size,
		Ids:     ids,
		Config:  rsCfg,
	}
	return &RSResumablePutStream{putStream, token}, nil
}

// CurrentSize IO: 请求数据服务器获取分片大小
func (p *RSResumablePutStream) CurrentSize() (int64, error) {
	//只请求一个服务器，因为Rs算法保证每次上传到每个服务器的大小一致
	size, err := webapi.HeadTmpObject(p.Servers[0], p.Ids[0])
	if err != nil {
		return 0, err
	}
	//求乘积得到当前大小
	size *= int64(p.rsConfig.DataShards)
	if size > p.Size {
		return p.Size, nil
	}
	return size, nil
}

// Token 上传记录
func (p *RSResumablePutStream) Token() string {
	tk := resumeToken{
		Name:    p.Name,
		Hash:    p.Hash,
		Size:    p.Size,
		Servers: p.Servers,
		Ids:     p.Ids,
		Config:  p.Config,
	}
	return base64.StdEncoding.EncodeToString(util.GobEncode(tk))
}
