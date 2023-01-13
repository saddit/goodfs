package slices

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSafeChunk(t *testing.T) {
	at := assert.New(t)
	s := []int{0, 1, 2, 3, 4, 5, 6}
	chunk := SafeChunk(s, 2, -2)
	at.Equal(4, len(chunk))
}

func TestNegMod(t *testing.T) {
	t.Log(-8 % -3)
	t.Log(8 % 3)
}

func TestClear(t *testing.T) {
	arr := []int{1, 2, 3, 4}
	Clear(&arr)
	assert.New(t).Equal(0, len(arr))
}
