package request

import (
	"common/util"
	"strings"
)

type Int64Tuple struct {
	First  int64
	Second int64
}

type Range struct {
	Bytes []Int64Tuple
}

//GetFirstBytes return first element
func (rg *Range) GetFirstBytes() (Int64Tuple, bool) {
	if rg.Bytes != nil && len(rg.Bytes) > 0 {
		return rg.Bytes[0], true
	}
	return Int64Tuple{}, false
}

//FirstBytes if contained, return first element else panic
func (rg *Range) FirstBytes() Int64Tuple {
	tp, _ := rg.GetFirstBytes()
	return tp
}

//ConvertFrom must start with bytes=
func (rg *Range) ConvertFrom(str string) bool {
	if str == "" {
		return false
	}
	_, str, ok := strings.Cut(str, "bytes=")
	if !ok {
		return false
	}
	tuples := strings.Split(str, ",")
	if len(tuples) == 0 {
		return false
	}
	rg.Bytes = make([]Int64Tuple, 0, len(tuples))
	for _, t := range tuples {
		v := strings.Split(strings.TrimSpace(t), "-")
		var tp Int64Tuple
		if len(v) > 0 {
			tp.First = util.ToInt64(v[0])
		}
		if len(v) > 1 {
			tp.Second = util.ToInt64(v[1])
		}
		rg.Bytes = append(rg.Bytes, tp)
	}
	return true
}
