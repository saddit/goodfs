package service

import (
	"metaserver/internal/entity"
	. "metaserver/internal/usecase"
)

type MetadataService struct {
	repo IMetadataRepo
}

func NewMetadataService(repo IMetadataRepo) *MetadataService {
	return &MetadataService{repo}
}

func (m *MetadataService) AddMetadata(name string, data *entity.Metadata) error {
	//TODO 添加元数据
	return nil
}

func (m *MetadataService) AddVersion(name string, data *entity.Version) (int, error) {
	//TODO 添加版本，需要原子操作确定版本号
	return -1, nil
}

func (m *MetadataService) UpdateMetadata(name string, data *entity.Metadata) error {
	//TODO 更新元数据
	return nil
}

func (m *MetadataService) UpdateVersion(name string, ver int, data *entity.Version) error {
	//TODO 更新版本数据
	return nil
}

func (m *MetadataService) RemoveMetadata(name string) error {
	//TODO 删除全部元数据
	return nil
}

func (m *MetadataService) RemoveVersion(name string, ver int) error {
	//TODO 删除版本数据
	return nil
}

func (m *MetadataService) GetMetadata(name string) (*entity.Metadata, error) {
	//TODO 获取元数据，无版本信息
	return nil, nil
}

func (m *MetadataService) GetVersion(name string, ver int) (*entity.Metadata, error) {
	//TODO 获取单个版本元数据
	return nil, nil
}

func (m *MetadataService) ListVersions(name string, start int, end int) ([]*entity.Version, error) {
	//TODO 获取版本数据集合
	return nil, nil
}
