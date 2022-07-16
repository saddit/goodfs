package webapi

import (
	"apiserver/internal/entity"
	"fmt"
	"io"
)

func GetMetadata(ip, name string) (*entity.Metadata, error) {
	//TODO 获取元数据（不包括版本）
	return nil, nil
}

func PostMetadata(ip, name string) error {
	//TODO 新增元数据请求（不包括版本）
	return nil
}

func PutMetadata(ip, name string) error {
	//TODO 覆盖元数据请求 （不包括版本元数据）
	return nil
}

func DelMetadata(ip, name string) error {
	//TODO 删除元数据请求（包括版本元数据）
	return nil
}

func GetVersion(ip, name string, verNum int) (*entity.Version, error) {
	//TODO 获取版本请求
	return nil, nil
}

func PostVersion(ip, name string, body io.Reader) (int, error) {
	//TODO 新增版本请求
	return -1, nil
}

func PutVersion(ip, name string, verNum int, body io.Reader) error {
	//TODO 覆盖版本请求
	return nil
}

func DelVersion(ip, name string, verNum int) error {
	//TODO 删除版本请求
	return nil
}

func metaRest(ip, name string) string {
	return fmt.Sprintf("http://%s/metadata/%s", ip, name)
}

func versionRest(ip, name string) string {
	return fmt.Sprintf("http://%s/metadata_version/%s", ip, name)
}
