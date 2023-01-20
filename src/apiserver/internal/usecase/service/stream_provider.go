package service

import (
	"apiserver/config"
)

type StreamOption struct {
	Locates  []string
	Bucket   string
	Hash     string
	Name     string
	Size     int64
	Compress bool
	Updater  LocatesUpdater
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
	}
}

type streamProvider struct {
	getStream func([]string) (ReadSeekCloser, error)
	puStream  func([]string) (WriteCommitCloser, error)
}

func (sp *streamProvider) GetStream(ips []string) (ReadSeekCloser, error) {
	return sp.getStream(ips)
}

func (sp *streamProvider) PutStream(ips []string) (WriteCommitCloser, error) {
	return sp.puStream(ips)
}
