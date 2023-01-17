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

// LimitReader returns a Reader that reads from r
// but stops with EOF after n bytes.
// The underlying implementation is a *LimitedReader.
func LimitReader(r io.Reader, n int64) io.Reader { return &LimitedReader{r, n} }

// A LimitedReader reads from R but limits the amount of
// data returned to just N bytes. Each call to Read
// updates N to reflect the new amount remaining.
// Read returns EOF when N <= 0 or when the underlying R returns EOF.
type LimitedReader struct {
	R io.Reader // underlying reader
	N int64  // max bytes remaining
}

func (l *LimitedReader) Read(p []byte) (n int, err error) {
	if l.N <= 0 {
		return 0, io.EOF
	}
	if int64(len(p)) > l.N {
		p = p[0:AlignedSize64(l.N)]
	}
	n, err = l.R.Read(p)
	if int64(n) > l.N {
		n = int(l.N)
	}
	l.N -= int64(n)
	return
}
