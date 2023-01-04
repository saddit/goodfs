package disk

import "io"

func LimitWriter(r io.Writer, n int64) io.Writer { return &LimitedWriter{r, n} }

type LimitedWriter struct {
	R io.Writer // underlying reader
	N int64     // max bytes remaining
}

func (l *LimitedWriter) Write(p []byte) (n int, err error) {
	if l.N <= 0 {
		return len(p), nil
	}
	ln := int64(len(p))
	if ln > l.N {
		p = p[0:l.N]
	}
	n, err = l.R.Write(p)
	l.N -= int64(n)
	if l.N == 0 && n == len(p) {
		n = int(ln)
	}
	return
}
