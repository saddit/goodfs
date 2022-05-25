package selector

import (
	"apiserver/internal/entity"
	"math/rand"
	"time"
)

const Random SelectStrategy = "random"

type RandomSelector struct{}

func (s *RandomSelector) Pop(ds []*entity.DataServ) ([]*entity.DataServ, string) {
	rand.Seed(time.Now().Unix())
	idx := rand.Intn(len(ds))
	ip := ds[idx].Ip
	ds[idx] = ds[len(ds)-1]
	return ds[:len(ds)-1], ip
}

func (s *RandomSelector) Strategy() SelectStrategy {
	return Random
}

func (s *RandomSelector) Select(ds []*entity.DataServ) string {
	rand.Seed(time.Now().Unix())
	return ds[rand.Intn(len(ds))].Ip
}
