package set

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOfMapKeys(t *testing.T) {
	mp := map[string]int{
		"A": 1,
		"B": 2,
	}
	st := OfMapKeys(mp)
	assert.New(t).True(st.Contains("A"))
	assert.New(t).True(st.Contains("B"))
	assert.New(t).False(st.Contains(1))
	assert.New(t).False(st.Contains(2))
}
