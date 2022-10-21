package entity

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

var FullBindings = []interface{}{binding.Uri, binding.Query, binding.JSON}

type Binder struct{}

func (b *Binder) Bind(c *gin.Context, json bool) error {
	if json {
		return BindAll(c, b, FullBindings...)
	}
	return BindAll(c, b, FullBindings[:len(FullBindings)-1]...)
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
