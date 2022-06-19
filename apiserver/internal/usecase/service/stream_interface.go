package service

import "io"

type Commiter interface {
	Commit(bool) error
}

type WriteCloseCommiter interface {
	io.WriteCloser
	Commiter
}

type WriteCommiter interface {
	io.Writer
	Commiter
}
