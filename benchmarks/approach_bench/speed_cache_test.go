package main_test

import (
	"testing"
)

const CACHE_LIMIT = 1000

var cache map[fraction][]int = make(map[fraction][]int)

func initCache() {
	f := fraction{1, 1}
	for ; f.a < CACHE_LIMIT; f.a++ {
		for ; f.b < CACHE_LIMIT; f.b++ {
			cache[f] = getContinuedFracCache(f)
		}
	}
}

func computeCache(weights *map[int]int, wrk worker) {
	var f fraction
	for _, rect := range wrk.assigned {
		for f.a = rect.xl; f.a <= rect.xr; f.a++ {
			for f.b = rect.yl; f.b <= rect.yr; f.b++ {
				for _, num := range getContinuedFracCache(f) {
					(*weights)[num]++
				}
			}
		}
	}
	wg.Done()
}

func getContinuedFracCache(f fraction) []int {
	var contFrac []int = make([]int, 0)
	for f.b > 0 {
		contFrac = append(contFrac, f.a/f.b)
		f.a, f.b = f.b, f.a%f.b

		if _, ok := cache[f]; ok {
			return append(contFrac, cache[f]...)
		}
	}
	return contFrac
}

func BenchmarkCache(b *testing.B) {
	initCache()
	b.ResetTimer()
	benchmark(computeCache, b)
}
