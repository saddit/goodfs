package repo

import "common/cache"

type MetadataCacheRepo struct {
	cache cache.ICache
}

func NewMetadataCacheRepo(c cache.ICache) *MetadataCacheRepo {
	return &MetadataCacheRepo{c}
}

//TODO cache operation