package command

import (
	"errors"
)

var (
	ErrNotSupport = errors.New("-r supports only exist_filter")
)

func RepairCommand(cmd Command) error {
	switch cmd.R {
	case "none":
		break
	default:
		return ErrNotSupport
	}
	return nil
}
