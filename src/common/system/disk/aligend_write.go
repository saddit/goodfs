package disk

import (
	"common/util/math"
	"errors"
	"io"

	"github.com/ncw/directio"
)

func AlignedSize(n int) int {
	if i := n % directio.BlockSize; i > 0 {
		return n - i + directio.BlockSize
	}
	return n
}

func AlignedSize64(n int64) int64 {
	if i := n % directio.BlockSize; i > 0 {
		return n - i + directio.BlockSize
	}
	return n
}

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
	n = math.MinInt(length, AlignedSize(n))
	return
}

// AlignedWriter impl io.ReaderFrom interface
// Write data aligned to multiple of 4KB
type AlignedWriter struct {
	io.Writer
}

func NewAlignedWriter(wt io.Writer) *AlignedWriter {
	return &AlignedWriter{wt}
}

func (aw *AlignedWriter) ReadFrom(rd io.Reader) (int64, error) {
	return AlignedWriteTo(aw, rd, 8*directio.BlockSize)
}

// AlignedWriteTo fill zero padding to multiple of 4KB if not enough
func AlignedWriteTo(dst io.Writer, src io.Reader, bufSize int) (written int64, err error) {
	buf := AlignedBlock(bufSize)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			var nw int
			var ew error
			if i := nr % directio.BlockSize; i > 0 {
				newBuf := AlignedBlock(nr - i + directio.BlockSize)
				copy(newBuf, buf[0:nr])
				nr = len(newBuf)
				nw, ew = dst.Write(newBuf)
			} else {
				nw, ew = dst.Write(buf[0:nr])
			}
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = errors.New("invalid write result")
				}
			}
			written += int64(nw)
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return written, err
}
