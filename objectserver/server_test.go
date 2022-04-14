package main

import "testing"

func TestKeyValue(t *testing.T) {
	mp := make(map[string]interface{})
	println(mp["Evict"].(bool))
}
