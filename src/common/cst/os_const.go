package cst

import (
	"io/fs"
	"os"
)

var OS = struct {
	ModeUser  fs.FileMode
	WriteFlag int
}{
	ModeUser:  0600,
	WriteFlag: os.O_WRONLY | os.O_CREATE | os.O_TRUNC,
}
