package service

import (
	"apiserver/internal/entity"
	"apiserver/internal/usecase"
	"apiserver/internal/usecase/pool"
	"apiserver/internal/usecase/repo"
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
	metaD, err := m.repo.FindByNameWithVersion(md.Name, entity.VerModeNot)
	if err != nil {
		return -1, nil
	}
	var verNum int32
	if metaD != nil {
		verNum, err = m.versionRepo.Add(metaD.Name, ver)
		if err != nil {
			return -1, err
		}
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

func (m *MetaService) UpdateVersion(name string, version *entity.Version) (err error) {
	err = m.versionRepo.Update(name, version)
	return
}

func (m *MetaService) GetVersion(name string, version int32) (*entity.Version, error) {
	res, err := m.versionRepo.Find(name, version)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, usecase.ErrNotFound
	}
	return res, nil
}

func (m *MetaService) GetMetadata(name string, ver int32) (*entity.Metadata, error) {
	verMode := entity.VerMode(ver)
	res, err := m.repo.FindByNameWithVersion(name, verMode)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, usecase.ErrNotFound
	}
	if verMode != entity.VerModeNot && len(res.Versions) == 0 {
		return res, usecase.ErrNotFound
	}
	return res, nil
}
