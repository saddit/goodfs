package utils

import (
	"github.com/sirupsen/logrus"
	"github.com/tinylib/msgp/msgp"
)

// decodeMsg decode data by msgp if error return false
func DecodeMsgp[T msgp.Unmarshaler](data T, bt []byte) bool {
	if _, err := data.UnmarshalMsg(bt); err != nil {
		logrus.Errorf("%T decode err: %v", data, err)
		return false
	}
	return true
}

// encodeMsg encode data with msgp if error return nil
func EncodeMsgp(data msgp.MarshalSizer) []byte {
	bt, err := data.MarshalMsg(nil)
	if err != nil {
		logrus.Errorf("%T encode err: %v", data, err)
		return nil
	}
	return bt
}