package main_test

import (
	"math"
	"strconv"
	"sync"
	"testing"
)

const (
	thousand = 1000

	MAX_CELLS = 640
	MIN_CELLS = 10
	MUL_CELLS = 4

	MAX_WORKERS = 4096
	MIN_WORKERS = 16
	MUL_WORKERS = 16
)

var wg sync.WaitGroup

type borders struct {
	xl, xr int
	yl, yr int
}

type worker struct {
	assigned []borders
}

type fraction struct {
	a, b int
}

func run(size, cells, workers int, compute func(*map[int]int, worker)) {
	wrkrs := initWorkers(size, cells, workers)
	engageWorkers(wrkrs, compute)
}

func initWorkers(size, cells, workers int) []worker {
	bounds := initBounds(size, cells)

	wrkrs := make([]worker, workers)
	for i := range wrkrs {
		wrkrs[i].assigned = make([]borders, 0)
	}

	if len(bounds) < len(wrkrs) {
		wrkrs = wrkrs[:len(bounds)]
	}

	for i, w := 0, 0; i < len(bounds); i, w = i+1, w+1 {
		if w == len(wrkrs) {
			w = 0
		}
		wrkrs[w].assigned = append(wrkrs[w].assigned, bounds[i])
	}

	return wrkrs
}

func engageWorkers(wrkrs []worker, compute func(*map[int]int, worker)) {
	weightArr := make([]map[int]int, len(wrkrs))
	for i := range weightArr {
		weightArr[i] = make(map[int]int)
	}

	wg.Add(len(wrkrs))
	for i := range wrkrs {
		go compute(&weightArr[i], wrkrs[i])
	}
	wg.Wait()
}

func initBounds(size, cells int) []borders {
	if size <= 1000 {
		return []borders{
			{
				xl: 1,
				yl: 1,
				xr: size,
				yr: size,
			},
		}
	}

	if size/cells < 1 {
		cells = size / 10
	}
	bounds := make([]borders, cells*cells)

	var x, y, i, step int = 1, 1, 0, int(math.Ceil(float64(size) / float64(cells)))
	for x = 1; x < size-(step+1); x += step + 1 {
		for y = 1; y < size-(step+1); y += step + 1 {
			bounds[i] = borders{
				xl: x,
				xr: x + step,
				yl: y,
				yr: y + step,
			}
			i++
		}
		bounds[i] = borders{
			xl: x,
			xr: x + step,
			yl: y,
			yr: size,
		}
		i++
	}

	for y = 1; y < size-(step+1); y += step + 1 {
		bounds[i] = borders{
			xl: x,
			xr: size,
			yl: y,
			yr: y + step,
		}
		i++
	}

	bounds[i] = borders{
		xl: x,
		xr: size,
		yl: y,
		yr: size,
	}
	i++

	if i < len(bounds)-1 {
		bounds = bounds[:i]
	}

	return bounds
}

func getContinuedFrac(f fraction) []int {
	var contFrac []int = make([]int, 0)
	for f.b > 0 {
		contFrac = append(contFrac, f.a/f.b)
		f.a, f.b = f.b, f.a%f.b
	}
	return contFrac
}

func benchmark(compute func(*map[int]int, worker), b *testing.B) {
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
	}

	var name string
	for _, test := range testset {
		for cells := MIN_CELLS; cells <= MAX_CELLS; cells *= MUL_CELLS {
			name = test.name + "_c" + strconv.Itoa(cells)
			for workers := MIN_WORKERS; workers <= MAX_WORKERS; workers *= MUL_WORKERS {
				fullname := name + "_w" + strconv.Itoa(workers)
				b.Run(fullname, func(b *testing.B) {
					for i := 0; i < b.N; i++ {
						run(test.size, cells, workers, compute)
					}
				})
			}
		}
	}
}
