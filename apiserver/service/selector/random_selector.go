package selector

import (
	"goodfs/apiserver/model/dataserv"
	"math/rand"
	"time"
)

const Random SelectStrategy = "random"

type RandomSelector struct{}

func (s *RandomSelector) Strategy() SelectStrategy {
	return Random
}

func (s *RandomSelector) Select(ds []*dataserv.DataServ) string {
	rand.Seed(time.Now().Unix())
	return ds[rand.Intn(len(ds))].Ip
}
