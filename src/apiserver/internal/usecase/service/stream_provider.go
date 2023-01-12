package service

import (
	"apiserver/config"
	"apiserver/internal/entity"
)

type StreamOption struct {
	Locates []string
	Hash    string
	Name    string
	Size    int64
	Updater LocatesUpdater
}

func RsStreamProvider(opt *StreamOption, cfg *config.RsConfig) StreamProvider {
	return &streamProvider{
		getStream: func(s []string) (ReadSeekCloser, error) {
			opt.Locates = s
			return NewRSGetStream(opt, cfg)
		},
		puStream: func(s []string) (WriteCommitCloser, error) {
			opt.Locates = s
			return NewRSPutStream(opt, cfg)
		},
		fillMeta: func(v *entity.Version) {
			v.DataShards = cfg.DataShards
			v.ParityShards = cfg.ParityShards
			v.ShardSize = cfg.ShardSize(v.Size)
		},
	}
}

func CpStreamProvider(opt *StreamOption, cfg *config.ReplicationConfig) StreamProvider {
	return &streamProvider{
		getStream: func(s []string) (ReadSeekCloser, error) {
			opt.Locates = s
			return NewCopyGetStream(opt, cfg)
		},
		puStream: func(s []string) (WriteCommitCloser, error) {
			opt.Locates = s
			return NewCopyPutStream(opt, cfg)
		},
		fillMeta: func(v *entity.Version) {
			v.DataShards = cfg.CopiesCount
			v.ShardSize = int(v.Size)
		},
	}
}

type streamProvider struct {
	getStream func([]string) (ReadSeekCloser, error)
	puStream  func([]string) (WriteCommitCloser, error)
	fillMeta  func(*entity.Version)
}

func (sp *streamProvider) GetStream(ips []string) (ReadSeekCloser, error) {
	return sp.getStream(ips)
}

func (sp *streamProvider) PutStream(ips []string) (WriteCommitCloser, error) {
	return sp.puStream(ips)
}

func (sp *streamProvider) FillMetadata(v *entity.Version) {
	sp.fillMeta(v)
}
