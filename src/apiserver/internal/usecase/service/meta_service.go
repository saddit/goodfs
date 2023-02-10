package service

import (
	"apiserver/internal/entity"
	"apiserver/internal/usecase"
	"apiserver/internal/usecase/repo"
)

type MetaService struct {
	repo        repo.IMetadataRepo
	versionRepo repo.IVersionRepo
}

func NewMetaService(repo repo.IMetadataRepo, versionRepo repo.IVersionRepo) *MetaService {
	return &MetaService{repo: repo, versionRepo: versionRepo}
}

func (m *MetaService) AddVersion(name, bucket string, version *entity.Version) (int32, error) {
	return m.versionRepo.Add(name, bucket, version)
}

func (m *MetaService) SaveMetadata(md *entity.Metadata) (int32, error) {
	if err := m.repo.Insert(md); err != nil {
		return 0, err
	}
	if len(md.Versions) > 0 {
		return m.AddVersion(md.Name, md.Bucket, md.Versions[0])
	}
	return 0, nil
}

func (m *MetaService) UpdateVersion(name, bucket string, version *entity.Version) (err error) {
	err = m.versionRepo.Update(name, bucket, version)
	return
}

func (m *MetaService) RemoveVersion(name, bucket string, version int32) error {
	return m.versionRepo.Delete(name, bucket, version)
}

func (m *MetaService) GetVersion(name, bucket string, version int32) (*entity.Version, error) {
	res, err := m.versionRepo.Find(name, bucket, version)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, usecase.ErrNotFound
	}
	return res, nil
}

func (m *MetaService) GetMetadata(name, bucket string, ver int32, withExtra bool) (*entity.Metadata, error) {
	verMode := entity.VerMode(ver)
	res, err := m.repo.FindByName(name, bucket, withExtra)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, usecase.ErrNotFound
	}
	if verMode != entity.VerModeNot {
		v, err := m.versionRepo.Find(name, bucket, ver)
		if err != nil {
			return nil, err
		}
		res.Versions = append(res.Versions, v)
	}
	return res, nil
}
