package service

import (
	"apiserver/internal/entity"
	"apiserver/internal/usecase"
	"apiserver/internal/usecase/pool"
	"apiserver/internal/usecase/repo"
	"context"
)

type MetaService struct {
	repo        repo.IMetadataRepo
	versionRepo repo.IVersionRepo
}

func NewMetaService(repo repo.IMetadataRepo, versionRepo repo.IVersionRepo) *MetaService {
	return &MetaService{repo: repo, versionRepo: versionRepo}
}

func saveAlgoInfo(ver *entity.Version) {
	ver.DataShards = pool.Config.Rs.DataShards
	ver.ParityShards = pool.Config.Rs.ParityShards
	ver.ShardSize = pool.Config.Rs.BlockPerShard
	ver.EcAlgo = 1
}

func (m *MetaService) SaveMetadata(md *entity.Metadata) (int32, error) {
	ver := md.Versions[0]
	saveAlgoInfo(ver)
	metaD := m.repo.FindByNameAndVerMode(md.Name, entity.VerModeNot)
	var verNum int32
	if metaD != nil {
		verNum = m.versionRepo.Add(context.Background(), metaD.Name, ver)
	} else {
		verNum = 0
		var e error
		if _, e = m.repo.Insert(md); e != nil {
			verNum = repo.ErrVersion
		}
	}

	if verNum == repo.ErrVersion {
		return -1, usecase.ErrInternalServer
	} else {
		return verNum, nil
	}
}

func (m *MetaService) UpdateVersion(version *entity.Version) {
	m.versionRepo.Update(context.Background(), version)
}

func (m *MetaService) GetVersion(name string, version int32) (*entity.Version, bool) {
	res := m.versionRepo.Find(name, version)
	if res == nil {
		return nil, false
	}
	return res, true
}

func (m *MetaService) GetMetadata(name string, ver int32) (*entity.Metadata, bool) {
	verMode := entity.VerMode(ver)
	res := m.repo.FindByNameAndVerMode(name, verMode)
	if res == nil {
		return nil, false
	}
	return res, verMode != entity.VerModeNot || len(res.Versions) > 0
}
