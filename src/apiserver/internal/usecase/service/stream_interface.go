package service

import "io"

type Committer interface {
	Commit(bool) error
}

type WriteCloseCommitter interface {
	io.WriteCloser
	Committer
}

type WriteCommitter interface {
	io.Writer
	Committer
}
