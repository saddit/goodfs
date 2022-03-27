package main

import (
	"goodfs/api"
	"goodfs/objects"
	"sync"
)

func main() {
	go objects.Start()
	go api.Start()

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
