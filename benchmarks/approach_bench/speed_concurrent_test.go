package main_test

import (
	"testing"
)

func computeConcurrent(weights *map[int]int, wrk worker) {
	var f fraction
	for _, rect := range wrk.assigned {
		for f.a = rect.xl; f.a <= rect.xr; f.a++ {
			for f.b = rect.yl; f.b <= rect.yr; f.b++ {
				for _, num := range getContinuedFrac(f) {
					(*weights)[num]++
				}
			}
		}
	}
	wg.Done()
}

func BenchmarkConcurrent(b *testing.B) {
	benchmark(computeConcurrent, b)
}
