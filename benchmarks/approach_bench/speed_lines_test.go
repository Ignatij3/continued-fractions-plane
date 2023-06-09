package main_test

import (
	"math"
	"strconv"
	"sync"
	"testing"
)

var maplock sync.RWMutex

func runLines(size, cells, workers int) {
	diagonal, wrkrs := initWorkersLines(size, cells, workers)
	engageWorkersLines(size, diagonal, wrkrs)
}

func initWorkersLines(size, cells, workers int) ([]borders, []worker) {
	diagonal, bounds := initBoundsLines(size, cells)

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

	return diagonal, wrkrs
}

func initBoundsLines(size, cells int) ([]borders, []borders) {
	if size <= 1000 {
		return []borders{
			{
				xl: 1,
				yl: 1,
				xr: size,
				yr: size,
			},
		}, []borders{}
	}

	if size/cells < 1 {
		cells = size / 10
	}

	step := int(math.Ceil(float64(size) / float64(cells)))
	diagonal := make([]borders, cells)
	bounds := make([]borders, cells*cells+cells/2)

	var a, i int = 1, 0
	for ; a < size-step; a += step + 1 {
		diagonal[i] = borders{
			xl: a,
			xr: a + step,
			yl: a,
			yr: a + step,
		}
		i++
	}
	diagonal[i] = borders{
		xl: a,
		xr: size,
		yl: a,
		yr: size,
	}
	i++

	if i < len(diagonal)-1 {
		diagonal = diagonal[:i]
	}
	i = 0

	var y int
	for j := 0; j < len(diagonal)-2; j++ {
		for y = diagonal[j].yr + 1; y < size-(step+1); y += step + 1 {
			bounds[i] = borders{
				xl: diagonal[j].xl,
				xr: diagonal[j].xr,
				yl: y,
				yr: y + step,
			}
			i++
		}
		bounds[i] = borders{
			xl: diagonal[j].xl,
			xr: diagonal[j].xr,
			yl: y,
			yr: size,
		}
		i++
	}

	bounds[i] = borders{
		xl: diagonal[len(diagonal)-2].xl,
		xr: diagonal[len(diagonal)-2].xr,
		yl: diagonal[len(diagonal)-2].yr + 1,
		yr: size,
	}
	i++

	if i < len(bounds)-1 {
		bounds = bounds[:i]
	}

	return diagonal, bounds
}

func engageWorkersLines(size int, diagonal []borders, wrkrs []worker) {
	weights := make(map[int]int)
	for key := 0; key <= size; key++ {
		weights[key] = 0
	}
	weights[1] = size

	wg.Add(len(wrkrs) + 1)
	for i := range wrkrs {
		go computeLines(size, &weights, wrkrs[i])
	}
	go computeDiagonal(size, &weights, worker{assigned: diagonal})
	wg.Wait()
}

func computeLines(size int, weights *map[int]int, wrk worker) {
	var (
		f        fraction
		fracAmt  int
		contFrac []int
	)

	for _, rect := range wrk.assigned {
		for a := rect.xl; a <= rect.xr; a++ {
			for b := rect.yl; b <= rect.yr; b++ {
				f.a, f.b = a, b
				if gcd(f.a, f.b) == 1 {
					if f.a < f.b {
						f.a, f.b = f.b, f.a
					}

					contFrac = getContinuedFrac(f)
					fracAmt = size / f.a

					maplock.Lock()
					(*weights)[0] += fracAmt // добавляется кол-во нулей
					for _, n := range contFrac {
						(*weights)[n] += fracAmt * 2 //считается и обратная дробь
					}
					maplock.Unlock()
				}
			}
		}
	}

	wg.Done()
}

func computeDiagonal(size int, weights *map[int]int, wrk worker) {
	var (
		f        fraction
		fracAmt  int
		contFrac []int
	)

	for _, rect := range wrk.assigned {
		for f.b = rect.xl; f.b <= rect.xr; f.b++ {
			for f.a = f.b + 1; f.a <= rect.yr; f.a++ {
				if gcd(f.a, f.b) == 1 {
					contFrac = getContinuedFrac(f)
					fracAmt = size / f.a

					maplock.Lock()
					for _, n := range contFrac {
						(*weights)[n] += fracAmt * 2
					}
					maplock.Unlock()
				}
			}
		}
	}

	wg.Done()
}

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func BenchmarkLines(b *testing.B) {
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
						runLines(test.size, cells, workers)
					}
				})
			}
		}
	}
}
