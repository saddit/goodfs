package model

import (
	"strconv"
	"strings"
)

type Int64Tuple struct {
	First  int64
	Second int64
}

type Range struct {
	Bytes []Int64Tuple
}

//Get return first element
func (rg *Range) Get() (Int64Tuple, bool) {
	if rg.Bytes != nil && len(rg.Bytes) > 0 {
		return rg.Bytes[0], true
	}
	return Int64Tuple{}, false
}

//Value if contained, return first element else panic
func (rg *Range) Value() Int64Tuple {
	if v, ok := rg.Get(); ok {
		return v
	}
	panic("No value contains in range")
}

func (rg *Range) convertFrom(str string) bool {
	if _, str, ok := strings.Cut(str, "bytes="); ok {
		if tuples := strings.Split(str, ","); len(tuples) > 0 {
			rg.Bytes = make([]Int64Tuple, 0, len(tuples))
			for _, t := range tuples {
				if v := strings.Split(t, "-"); len(v) > 0 {
					var tp Int64Tuple
					tp.First, _ = strconv.ParseInt(v[0], 10, 0)
					if len(v) > 1 {
						tp.Second, _ = strconv.ParseInt(v[1], 10, 0)
					}
					rg.Bytes = append(rg.Bytes, tp)
				}
			}
			return true
		}
	}
	return false
}
