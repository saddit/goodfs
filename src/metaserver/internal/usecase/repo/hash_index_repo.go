package repo

import (
	"metaserver/internal/usecase/db"
	"metaserver/internal/usecase/logic"
)

type HashIndexRepo struct {
	Storage *db.Storage
}

func NewHashIndexRepo(s *db.Storage) *HashIndexRepo {
	return &HashIndexRepo{Storage: s}
}

func (h *HashIndexRepo) Remove(hash, key string) error {
	return h.Storage.Batch(logic.NewHashIndexLogic().RemoveIndex(hash, key))
}

func (h *HashIndexRepo) FindAll(hash string) (keys []string, err error) {
	err = h.Storage.View(logic.NewHashIndexLogic().GetIndex(hash, &keys))
	return
}

func (h *HashIndexRepo) Sync() error {
	return h.Storage.DB().Sync()
}
