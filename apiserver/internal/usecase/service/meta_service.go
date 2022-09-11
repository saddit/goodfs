package service

import (
	"apiserver/internal/entity"
	"apiserver/internal/usecase"
	"apiserver/internal/usecase/pool"
	"apiserver/internal/usecase/repo"
	"common/response"
	"net/http"
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
		// if err is NotFound error
		if respErr, ok := err.(response.IResponseErr); ok && respErr.GetStatus() == http.StatusNotFound {
			if _, err = m.repo.Insert(md); err != nil {
				return 0, err
			}
		} else {
			return 0, err
		}
	} else {
		_, err = m.versionRepo.Add(metaD.Name, ver)
		if err != nil {
			return 0, err
		}
	}

	return ver.Sequence, nil
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
