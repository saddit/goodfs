package datasize

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParse(t *testing.T) {
	i := MustParse("1000")
	assert.New(t).Equal(1000, int(i))
	i = MustParse("1000B")
	assert.New(t).Equal(1000, int(i))
	i = MustParse("10KB")
	assert.New(t).Equal(10<<10, int(i))
	i = MustParse("10MB")
	assert.New(t).Equal(10<<20, int(i))
	i = MustParse("10GB")
	assert.New(t).Equal(10<<30, int(i))
	i = MustParse("10TB")
	assert.New(t).Equal(10<<40, int(i))
	i = MustParse("1PB")
	assert.New(t).Equal(1<<50, int(i))
}
