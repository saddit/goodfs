package disk

import (
	"io"
)

type eofReader struct{}

func (eofReader) Read([]byte) (int, error) {
	return 0, io.EOF
}

type multiReader struct {
	readers []io.Reader
}

// MultiReader only difference with io.MultiReader is that 
// this Reader will read until the buffer is full or there are no Readers left (returns io.EOF)
func MultiReader(readers ...io.Reader) io.Reader {
	mr := &multiReader{readers: make([]io.Reader, len(readers))}
	copy(mr.readers, readers)
	return mr
}

func (mr *multiReader) Read(p []byte) (n int, err error) {
	bufSize := len(p)
	for len(mr.readers) > 0 {
		// Optimization to flatten nested multiReaders (Issue 13558).
		if len(mr.readers) == 1 {
			if r, ok := mr.readers[0].(*multiReader); ok {
				mr.readers = r.readers
				continue
			}
		}
		var nr int
		nr, err = mr.readers[0].Read(p[n:])
		n += nr
		if err == io.EOF {
			// Use eofReader instead of nil to avoid nil panic
			// after performing flatten (Issue 18232).
			mr.readers[0] = eofReader{} // permit earlier GC
			mr.readers = mr.readers[1:]
		}
		if n >= bufSize {
			if err == io.EOF && len(mr.readers) > 0 {
				// Don't return EOF yet. More readers remain.
				err = nil
			}
			return
		}
	}
	return n, io.EOF
}