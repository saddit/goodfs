package selector

import "goodfs/api/model/dataserv"

type MaxFreeDiskSelector struct{}

const MaxFreeDisk SelectStrategy = "maxfreedisk"

func (s *MaxFreeDiskSelector) Select(ds []*dataserv.DataServ) string {
	return ds[0].Ip
}

func (s *MaxFreeDiskSelector) Strategy() SelectStrategy {
	return MaxFreeDisk
}
