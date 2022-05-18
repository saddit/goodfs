package main_test

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func changeArray(arr []string) {
	arr[0] = "Changed!!"
}

func TestFuzz(t *testing.T) {
	s := make([]string, 1)
	s[0] = "Original"
	changeArray(s)
	assert.New(t).Equal("Changed!!", s[0])
}
