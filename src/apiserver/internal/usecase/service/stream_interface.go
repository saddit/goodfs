package service

import "io"

type Committer interface {
	Commit(bool) error
}

type WriteCommitCloser interface {
	io.WriteCloser
	Committer
}

type WriteCommitter interface {
	io.Writer
	Committer
}

type LocatesUpdater func(locates []string) error

type ReadSeekCloser interface {
	io.ReadSeekCloser
}

type StreamProvider interface {
	GetStream(ip []string) (ReadSeekCloser, error)
	PutStream(ip []string) (WriteCommitCloser, error)
}
