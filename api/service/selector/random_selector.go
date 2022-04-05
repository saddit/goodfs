package selector

import (
	"goodfs/api/model/dataserv"
	"math/rand"
)

const Random SelectStrategy = "random"

type RandomSelector struct{}

func (s *RandomSelector) Strategy() SelectStrategy {
	return Random
}

func (s *RandomSelector) Select(ds []*dataserv.DataServ) string {
	return ds[rand.Intn(len(ds))].Ip
}
