package service

import (
	"apiserver/config"
	"apiserver/internal/entity"
	"io"
)

func RsStreamProivder(meta *entity.Version, cfg *config.RsConfig) StreamProvider {
	return &streamProvider{
		getStream: func(s []string) (io.ReadSeekCloser, error) {
			return NewRSGetStream(meta.Size, meta.Hash, s, cfg)
		},
		puStream: func(s []string) (WriteCloseCommitter, error) {
			return NewRSPutStream(s, meta.Hash, meta.Size, cfg)
		},
	}
}

func CpStreamProvider(meta *entity.Version, cfg *config.ReplicationConfig) StreamProvider {
	return &streamProvider{
		getStream: func(s []string) (io.ReadSeekCloser, error) {
			return NewCopyGetStream(meta.Hash, s, meta.Size, cfg)
		},
		puStream: func(s []string) (WriteCloseCommitter, error) {
			return NewCopyPutStream(meta.Hash, meta.Size, s, cfg)
		},
	}
}

type streamProvider struct {
	getStream func([]string) (io.ReadSeekCloser, error)
	puStream  func([]string) (WriteCloseCommitter, error)
}

func (sp *streamProvider) GetStream(ips []string) (io.ReadSeekCloser, error) {
	return sp.getStream(ips)
}

func (sp *streamProvider) PutStream(ips []string) (WriteCloseCommitter, error) {
	return sp.puStream(ips)
}