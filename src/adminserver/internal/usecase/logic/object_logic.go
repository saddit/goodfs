package logic

import (
	"adminserver/internal/usecase/webapi"
	"common/util"
	"common/util/crypto"
	"io"
	"mime/multipart"
)

type Objects struct {
}

func NewObjects() *Objects {
	return &Objects{}
}

func (Objects) Upload(file *multipart.FileHeader) error {
	// open and checksum
	temp, err := file.Open()
	if err != nil {
		return err
	}
	hash := crypto.SHA256IO(temp)
	util.LogErr(temp.Close())
	// open and send request
	fileBody, err := file.Open()
	if err != nil {
		return err
	}
	return webapi.PutObjects(SelectApiServer(), file.Filename, hash, fileBody)
}

func (Objects) Download(name string, version int) (io.ReadCloser, error) {
	return webapi.GetObjects(SelectApiServer(), name, version)
}
