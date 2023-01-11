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

func RsStreamProvider(meta *entity.Version, updater LocatesUpdater, cfg *config.RsConfig) StreamProvider {
	return &streamProvider{
		getStream: func(s []string) (ReadSeekCloser, error) {
			return NewRSGetStream(&StreamOption{
				Locates: s,
				Hash:    meta.Hash,
				Size:    meta.Size,
				Updater: updater,
			}, cfg)
		},
		puStream: func(s []string) (WriteCommitCloser, error) {
			return NewRSPutStream(&StreamOption{
				Locates: s,
				Hash:    meta.Hash,
				Size:    meta.Size,
			}, cfg)
		},
	}
}

func CpStreamProvider(meta *entity.Version, updater LocatesUpdater, cfg *config.ReplicationConfig) StreamProvider {
	return &streamProvider{
		getStream: func(s []string) (ReadSeekCloser, error) {
			return NewCopyGetStream(&StreamOption{
				Locates: s,
				Hash:    meta.Hash,
				Size:    meta.Size,
				Updater: updater,
			}, cfg)
		},
		puStream: func(s []string) (WriteCommitCloser, error) {
			return NewCopyPutStream(&StreamOption{
				Locates: s,
				Hash:    meta.Hash,
				Size:    meta.Size,
			}, cfg)
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
