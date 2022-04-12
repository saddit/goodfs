package util

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"strings"

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
