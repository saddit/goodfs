package auth

import "common/logs"

type DisabledValidator struct {

}

func (dv *DisabledValidator) Verify(_ Credential) error {
	logs.Std().Warn("you doesn't turn on verification, this will cause security risks")
	return nil
}