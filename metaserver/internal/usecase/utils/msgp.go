package utils

import (
	"common/logs"
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsgp decode data by msgp if error return false
func DecodeMsgp[T msgp.Unmarshaler](data T, bt []byte) bool {
	if _, err := data.UnmarshalMsg(bt); err != nil {
		logs.Std().Errorf("%T decode err: %v", data, err)
		return false
	}
	return true
}

// EncodeMsgp encode data with msgp if error return nil
func EncodeMsgp(data msgp.MarshalSizer) []byte {
	bt, err := data.MarshalMsg(nil)
	if err != nil {
		logs.Std().Errorf("%T encode err: %v", data, err)
		return nil
	}
	return bt
}
