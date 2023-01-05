//go:build linux || netbsd || freebsd
// +build linux netbsd freebsd

// Copyright (c) 2015-2021 MinIO, Inc.
//
// This file is part of MinIO Object Storage stack
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package disk

import (
	"errors"
	"io"
	"os"
	"syscall"

	"github.com/ncw/directio"
	"golang.org/x/sys/unix"
)

// OpenFileDirectIO - bypass kernel cache.
func OpenFileDirectIO(filePath string, flag int, perm os.FileMode) (*os.File, error) {
	return directio.OpenFile(filePath, flag, perm)
}

// DisableDirectIO - disables directio mode.
func DisableDirectIO(f *os.File) error {
	fd := f.Fd()
	flag, err := unix.FcntlInt(fd, unix.F_GETFL, 0)
	if err != nil {
		return err
	}
	flag &= ^(syscall.O_DIRECT)
	_, err = unix.FcntlInt(fd, unix.F_SETFL, flag)
	return err
}

// AlignedBlock - pass through to directio implementation.
func AlignedBlock(blockSize int) []byte {
	return directio.AlignedBlock(blockSize)
}

// AligendWriteTo fill zero to multiple of 4KB if not enough
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