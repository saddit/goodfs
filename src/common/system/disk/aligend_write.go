package disk

import (
	"errors"
	"io"

	"github.com/ncw/directio"
)

// AligendWriter impl io.ReaderFrom interface
// Write data aligen to multiple of 4KB 
type AligendWriter struct {
	io.Writer
}

func NewAligendWriter(wt io.Writer) *AligendWriter {
	return &AligendWriter{wt}
}

func (aw *AligendWriter) ReadFrom(rd io.Reader) (int64, error) {
	return AligendWriteTo(aw, rd, 8 * directio.BlockSize)
}

// AligendWriteTo fill zero padding to multiple of 4KB if not enough
func AligendWriteTo(dst io.Writer, src io.Reader, bufSize int) (written int64, err error) {
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
