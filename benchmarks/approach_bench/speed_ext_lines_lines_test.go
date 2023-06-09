package main_test

import (
	"math"
	"strconv"
	"sync"
	"testing"
	"time"
)

var syncsync *sync.WaitGroup = &sync.WaitGroup{}

func runExtLinesLines(size, workers int) {
	weights := make(map[int]int)
	reschan := getResultChan(size, workers, &weights)
	processResults((workers/10)+1, &weights, reschan)
}

func getResultChan(size, workers int, weights *map[int]int) chan map[int]int {
	(*weights)[1] = int(float64(size) / math.Sqrt2)

	reschan := make(chan map[int]int, workers)
	jobs := getJobChan(size, workers)

	ressync.Add(int(workers))
	linelock := &sync.Mutex{}
	for i := 0; i < workers; i++ {
		go computeLLines(i+1, workers, ressync, linelock, size, jobs, reschan)
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

func getJobChan(size, workers int) chan int {
	lineID := make(chan int, workers)
	go func() {
		for id := 1; id <= size; id++ {
			lineID <- id
		}
		close(lineID)
	}()
	return lineID
}

func computeLLines(id, workers int, ressync *sync.WaitGroup, linelock *sync.Mutex, size int, jobs chan int, reschan chan map[int]int) {
	defer ressync.Done()
	var (
		cache [CACHESIZE]int
		res   *map[int]int = &map[int]int{}
	)

	cooldown := time.NewTimer(time.Second)

	for line := range jobs {
		select {
		case <-cooldown.C:
			syncsync.Wait()
			reschan <- *res
			res = &map[int]int{}
		default:
			processLine(size, line, &cache, res)
		}
	}

	reschan <- *res
	reschan <- flushCache(&cache)
}

func processLine(radius, y int, cache *[CACHESIZE]int, res *map[int]int) int {
	rightBoundSquared := radius*radius - y*y
	if y <= int(float64(radius)/math.Sqrt2) {
		rightBoundSquared = (y - 1) * (y - 1)
	}

	var x, xSqr int
	for ; xSqr <= rightBoundSquared; x++ {
		if gcd(x, y) == 1 {
			recordTermsLines(radius, fraction{a: y, b: x}, cache, res)
		}
		xSqr += x<<1 + 1
	}

	return x
}

func recordTermsLines(radius int, f fraction, cache *[CACHESIZE]int, data *map[int]int) {
	contFrac := getContinuedFrac(f)
	fracAmt := int(math.Sqrt(float64((radius * radius)) / float64((f.a*f.a + f.b*f.b))))

	for _, n := range contFrac {
		if n < CACHESIZE {
			cache[n] += fracAmt * 2
		} else {
			(*data)[n] += fracAmt * 2
		}
	}
}

func BenchmarkExtLinesLines(b *testing.B) {
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
		{
			name: "3e4",
			size: thousand * 30,
		},
	}

	for _, test := range testset {
		for workers := MIN_WORKERS; workers <= MAX_WORKERS; workers *= MUL_WORKERS {
			fullname := test.name + "w" + strconv.Itoa(workers)
			b.Run(fullname, func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					runExtLinesLines(test.size, workers)
				}
			})
		}
	}
}
