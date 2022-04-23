package command

import (
	"errors"
	"goodfs/apiserver/service/tool"
)

var (
	ErrMissIp     = errors.New("-r exist_filter require -I ip:port")
	ErrNotSupport = errors.New("-r supports only exist_filter")
)

func RepairCommand(cmd Command) error {
	switch cmd.R {
	case "exist_filter":
		if cmd.A == "" {
			return ErrMissIp
		}
		tool.RepairExistFilter(cmd.A)
		break
	case "none":
		break
	default:
		return ErrNotSupport
	}
	return nil
}
