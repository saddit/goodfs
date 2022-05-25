package selector

import "apiserver/internal/entity"

type MaxFreeDiskSelector struct{}

const MaxFreeDisk SelectStrategy = "maxfreedisk"

func (s *MaxFreeDiskSelector) Pop(ds []*entity.DataServ) ([]*entity.DataServ, string) {
	return ds[1:], ds[0].Ip
}

func (s *MaxFreeDiskSelector) Select(ds []*entity.DataServ) string {
	return ds[0].Ip
}

func (s *MaxFreeDiskSelector) Strategy() SelectStrategy {
	return MaxFreeDisk
}
