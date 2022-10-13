package constrant

import "io/fs"

var OS = struct {
	ModeUser fs.FileMode
} {
	ModeUser: 0700,
}