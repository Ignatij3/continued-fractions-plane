package main_test

import (
	"math"
	"strconv"
	"sync"
	"testing"
)

var ressync *sync.WaitGroup = &sync.WaitGroup{}

func runExtLinesSquares(size, cells, workers int) {
	weights := make(map[int]int)
	diagonal := make(chan borders, cells)
	rects := make(chan borders, workers)

	reschan := distributeJobs(size, cells, workers, diagonal, rects, &weights)
	processResults(cells, &weights, reschan)
}

func initJobs(size, cells int, diagonal, rects chan borders) {
	if size <= 1000 {
		diagonal <- borders{
			xl: 1,
			yl: 1,
			xr: size,
			yr: size,
		}
		close(diagonal)
		close(rects)
		return
	}

	if size/cells < 1 {
		cells = size / 10
	}
	step := int(math.Ceil(float64(size) / float64(cells)))

	go initDiagonal(diagonal, step, size)
	go initAboveDiagonal(rects, step, size)
}

func initDiagonal(diagonal chan borders, step, size int) {
	var a int = 1
	for ; a < size-step; a += step + 1 {
		diagonal <- borders{
			xl: a,
			xr: a + step,
			yl: a,
			yr: a + step,
		}
	}

	if a < size {
		diagonal <- borders{
			xl: a,
			xr: size,
			yl: a,
			yr: size,
		}
	}

	close(diagonal)
}

func initAboveDiagonal(rects chan borders, step, size int) {
	var y int = 1
	for x := 1; x < size-(step+1); x += step + 1 {
		for y = x + step + 1; y < size-(step+1); y += step + 1 {
			rects <- borders{
				xl: x,
				xr: x + step,
				yl: y,
				yr: y + step,
			}
		}
		rects <- borders{
			xl: x,
			xr: x + step,
			yl: y,
			yr: size,
		}
	}
	close(rects)
}

func distributeJobs(size, cells, workers int, diagonal, rects chan borders, weights *map[int]int) chan map[int]int {
	(*weights)[1] = size
	reschan := make(chan map[int]int, workers/10)
	initJobs(size, cells, diagonal, rects)

	ressync.Add(int(workers))
	go computeSquares(size, diagonal, reschan, onDiagonal)
	for i := 1; i < workers; i++ {
		go computeSquares(size, rects, reschan, aboveDiagonal)
	}

	defer func() {
		go func() {
			ressync.Wait()
			for len(reschan) != 0 {
			}
			close(reschan)
		}()
	}()

	return reschan
}

func computeSquares(size int, jobs chan borders, reschan chan map[int]int, traverseRect func(int, borders, *[CACHESIZE]int, *map[int]int)) {
	defer ressync.Done()
	var cache [CACHESIZE]int

	for rect := range jobs {
		res := &map[int]int{}
		traverseRect(size, rect, &cache, res)
		reschan <- *res
	}

	reschan <- flushCache(&cache)
}

func aboveDiagonal(size int, rect borders, cache *[CACHESIZE]int, res *map[int]int) {
	for b := rect.xl; b <= rect.xr; b++ {
		for a := rect.yl; a <= rect.yr; a++ {
			if gcd(a, b) == 1 {
				recordTermsSquares(size, fraction{a: a, b: b}, cache, res)
			}
		}
	}
}

func onDiagonal(size int, rect borders, cache *[CACHESIZE]int, res *map[int]int) {
	for b := rect.xl; b <= rect.xr; b++ {
		for a := b + 1; a <= rect.yr; a++ {
			if gcd(a, b) == 1 {
				recordTermsSquares(size, fraction{a: a, b: b}, cache, res)
			}
		}
	}
}

func recordTermsSquares(size int, f fraction, cache *[CACHESIZE]int, data *map[int]int) {
	contFrac := getContinuedFrac(f)
	fracAmt := size / f.a

	for _, n := range contFrac {
		if n < CACHESIZE {
			cache[n] += fracAmt * 2
		} else {
			(*data)[n] += fracAmt * 2
		}
	}
}

func BenchmarkExtLinesSquares(b *testing.B) {
	var testset = []struct {
		name string
		size int
	}{
		{
			name: "1e2",
			size: thousand / 10,
		},
		{
			name: "1e3",
			size: thousand,
		},
		{
			name: "1e4",
			size: thousand * 10,
		},
		{
			name: "2.2e4",
			size: thousand * 22,
		},
	}

	var name string
	for _, test := range testset {
		for cells := MIN_CELLS; cells <= MAX_CELLS; cells *= MUL_CELLS {
			name = test.name + "_c" + strconv.Itoa(cells)
			for workers := MIN_WORKERS; workers <= MAX_WORKERS; workers *= MUL_WORKERS {
				fullname := name + "_w" + strconv.Itoa(workers)
				b.Run(fullname, func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						runExtLinesSquares(test.size, cells, workers)
					}
				})
			}
		}
	}
}
