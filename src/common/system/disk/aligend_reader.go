package disk

import (
	"common/util/math"
	"common/util/slices"
	"io"

	"github.com/ncw/directio"
)

// AlignedSize aligned n to multiple of 4096
func AlignedSize(n int) int {
	if i := n % directio.BlockSize; i > 0 {
		return n - i + directio.BlockSize
	}
	return n
}

// AlignedSize64 is same as AlignedSize but returns int64 value
func AlignedSize64(n int64) int64 {
	if i := n % directio.BlockSize; i > 0 {
		return n - i + directio.BlockSize
	}
	return n
}

// AlignedReader reads from underlying reader, padding their size to multiple of 4096 if buffer is large enough
type AlignedReader struct {
	io.Reader
}

func NewAlignedReader(rd io.Reader) *AlignedReader {
	return &AlignedReader{rd}
}

func (ar *AlignedReader) Read(p []byte) (n int, err error) {
	length := len(p)
	n, err = io.ReadFull(ar.Reader, p)
	if err == io.ErrUnexpectedEOF {
		err = io.EOF
	}
	// padding zero if needed
	paddingEnd := math.MinInt(length, AlignedSize(n))
	slices.Fill(p[n:paddingEnd], 0)
	return paddingEnd, err
}

type AlignedLimitedReader struct {
	R io.Reader // underlying reader
	N int64     // max bytes remaining
}

func (l *AlignedLimitedReader) Read(p []byte) (n int, err error) {
	if l.N <= 0 {
		return 0, io.EOF
	}
	// aligned and reduce buffer for last read
	if int64(len(p)) > l.N {
		p = p[0:AlignedSize64(l.N)]
	}
	n, err = l.R.Read(p)
	// ignore aligned part and return actual size of data on last reading
	if int64(n) > l.N {
		n = int(l.N)
	}
	l.N -= int64(n)
	return
}
