package main_test

import (
	"sync"
)

const CACHESIZE = 10 + 1

var processingEnd *sync.WaitGroup = &sync.WaitGroup{}

func processResults(cacheAmt int, weights *map[int]int, reschan chan map[int]int) {
	processingEnd.Add(1)
	defer processingEnd.Done()

	cachesync := &sync.WaitGroup{}
	cachesync.Add(cacheAmt)
	cacheDrain := make(chan map[int]int, cacheAmt)

	go func() {
		cachesync.Wait()
		close(cacheDrain)
	}()

	for i := 0; i < cacheAmt; i++ {
		go processToCache(reschan, cacheDrain, cachesync)
	}

	for res := range reschan {
		for key, value := range res {
			(*weights)[key] += value
		}
	}

	for cachedata := range cacheDrain {
		for key, value := range cachedata {
			(*weights)[key] += value
		}
	}
}

func processToCache(reschan, cacheDrain chan map[int]int, cachesync *sync.WaitGroup) {
	defer cachesync.Done()
	cache := make(map[int]int)
	for res := range reschan {
		for key, value := range res {
			cache[key] += value
		}
	}
	cacheDrain <- cache
}

func flushCache(cache *[CACHESIZE]int) map[int]int {
	res := map[int]int{}
	for key := 1; key < CACHESIZE; key++ {
		res[key] = cache[key]
		cache[key] = 0
	}
	return res
}
