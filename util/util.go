package util

import (
	"crypto/sha256"
	"encoding/base64"
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetObjectID(id string) primitive.ObjectID {
	res, e := primitive.ObjectIDFromHex(id)
	if e == nil {
		return res
	}
	return primitive.NilObjectID
}

func GetFileExt(fileName string, withDot bool) (string, bool) {
	idx := strings.LastIndex(fileName, ".")
	if idx == -1 {
		return "", false
	}
	if !withDot {
		idx++
	}
	return fileName[idx:], true
}

func BindAll(c *gin.Context, obj interface{}, bindings ...interface{}) error {
	var e error
	for _, b := range bindings {
		if _, ok := b.(binding.BindingUri); ok {
			e = c.ShouldBindUri(obj)
		} else if trans, ok := b.(binding.BindingBody); ok {
			e = c.ShouldBindBodyWith(obj, trans)
		} else if trans2, ok := b.(binding.Binding); ok {
			e = c.ShouldBindWith(obj, trans2)
		}
	}
	return e
}

//SHA256Hash
func SHA256Hash(reader io.Reader) string {
	cryt := sha256.New()
	if _, e := io.CopyBuffer(cryt, reader, make([]byte, 2048)); e == nil {
		b := cryt.Sum(make([]byte, 0, cryt.Size()))
		return base64.StdEncoding.EncodeToString(b)
	}
	return ""
}
