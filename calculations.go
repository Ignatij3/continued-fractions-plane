package main

import (
	"context"
	"math"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"
	"time"
)

// TODO fix comments and logs

// sets cache for numbers [0; CACHESIZE-1].
const CACHESIZE = 10 + 1

// fraction represents fraction a/b, where a,b ∈ ℕ.
type fraction struct {
	a, b uint
}

// run sets up workers and distributes jobs to them, run finishes running when all results have been processed.
func (p *program) run() {
	defer func() {
		if err := recover(); err != nil {
			logger.Printf("FATAL: panic occured, shutting down unexpectedly: %v\n", err)
			p.updateState()
			logger.Print(debug.Stack())
			os.Exit(1)
		}
	}()

	exit, triggerExit := context.WithCancel(context.Background())
	finish := make(chan struct{})
	cleanupSync := &sync.WaitGroup{}
	cleanupSync.Add(1)

	go func() {
		logger.Println("INFO: Setting up state monitoring")

		termination := make(chan os.Signal, 1)
		signal.Notify(termination, os.Interrupt, syscall.SIGTERM)

	outer:
		for {
			select {
			case term := <-termination:
				logger.Printf("INFO: Process has been interrupted with %v, cleaning up\n", term)
				triggerExit()
				<-finish
				p.updateState()
				break outer

			case <-finish:
				p.clearFiles()
				if err := p.saveFinalResults(); err != nil {
					logger.Fatalf("FATAL: Couldn't write final results to file: %v\n Obtained data: %v\n", err, p.weights)
				}
				break outer
			}
		}

		cleanupSync.Done()
		close(finish)
	}()

	reschan := p.getResultChan(exit)
	p.processResults(reschan)

	finish <- struct{}{}
	cleanupSync.Wait()
}

// processResults receives term weights through reschan and adds them to underlying array.
func (p *program) processResults(reschan chan map[uint]uint64) {
	cacheAmt := int(p.WORKERS/10) + 1

	cachesync := &sync.WaitGroup{}
	cachesync.Add(cacheAmt)
	cacheDrain := make(chan map[uint]uint64, cacheAmt)

	go func() {
		cachesync.Wait()
		close(cacheDrain)
	}()

	for i := 0; i < cacheAmt; i++ {
		go p.processToCache(reschan, cacheDrain, cachesync)
	}

	for cachedata := range cacheDrain {
		for key, value := range cachedata {
			p.weights[key] += uint64(value)
		}
	}
}

// processToCache receives results through reschan and sends them back through cacheDrain on completion.
func (p *program) processToCache(reschan, cacheDrain chan map[uint]uint64, cachesync *sync.WaitGroup) {
	defer cachesync.Done()
	cache := make(map[uint]uint64)
	for res := range reschan {
		for key, value := range res {
			cache[key] += value
		}
	}
	cacheDrain <- cache
}

// distributeJobs distributes rectangles among workers and returns channel where the results would be sent.
func (p *program) getResultChan(exit context.Context) chan map[uint]uint64 {
	logger.Println("INFO: Initializing and distributing jobs")

	if p.weights == nil {
		p.weights = make([]uint64, p.N+1)
		p.weights[1] = uint64(float64(p.N) / math.Sqrt2)
	}

	reschan := make(chan map[uint]uint64, p.WORKERS)
	jobs := p.getJobChan()

	ressync := &sync.WaitGroup{}
	ressync.Add(int(p.WORKERS))

	linelock := &sync.Mutex{}
	for i := uint(0); i < p.WORKERS; i++ {
		go p.compute(ressync, linelock, jobs, reschan, exit)
	}

	go func() {
		ressync.Wait()
		for len(reschan) != 0 {
		}
		logger.Println("INFO: Closing results channel")
		close(reschan)
	}()

	return reschan
}

// initJobs passes rectangles that need to be processed through diagonal and rects channels.
// Rects that share diagonal with N×N plane are sent to channel diagonal,
// any other rects that lie strictly above diagonal are sent to channel rects.
func (p *program) getJobChan() chan uint {
	logger.Println("INFO: Initializing lines")

	lineID := make(chan uint, p.WORKERS)
	go func() {
		for id := p.LastLine + 1; id <= p.N; id++ {
			lineID <- id
		}
		close(lineID)
	}()
	return lineID
}

// computeLines counts continued fraction terms in rectangles above the diagonal of the plane N×N.
// It counts terms of irreducible fractions number of times fractions are present in the plane N×N, it also takes into account that the reverse of a/b (where a > b)
// has the same terms as a/b, but with the leading zero.
func (p *program) compute(ressync *sync.WaitGroup, linelock *sync.Mutex, jobs chan uint, reschan chan map[uint]uint64, exit context.Context) {
	defer ressync.Done()
	var (
		cache [CACHESIZE]uint64
		res   *map[uint]uint64 = &map[uint]uint64{}
	)

	cooldown := time.NewTimer(time.Minute)

work:
	for line := range jobs {
		select {
		case <-cooldown.C:
			reschan <- *res
			res = &map[uint]uint64{}

		case <-exit.Done():
			break work

		default:
			processLine(p.N, line, &cache, res)
			if line >= p.LastLine {
				linelock.Lock()
				p.LastLine = line
				linelock.Unlock()
			}
		}
	}

	reschan <- *res
	reschan <- p.flushCache(&cache)
}

// aboveDiagonal counts continued fraction terms above the diagonal of the plane N×N, the jobs's squares must be strictly above the diagonal (not touching it) of the plane.
func processLine(radius, y uint, cache *[CACHESIZE]uint64, res *map[uint]uint64) uint {
	rightBoundSquared := radius*radius - y*y
	if y <= uint(float64(radius)/math.Sqrt2) {
		rightBoundSquared = (y - 1) * (y - 1)
	}

	var x, xSqr uint
	for ; xSqr <= rightBoundSquared; x++ {
		if gcd(x, y) == 1 {
			recordTerms(radius, fraction{a: y, b: x}, cache, res)
		}
		xSqr += x<<1 + 1
	}

	return x
}

// recordTerms updates cache and data map with information about continued fraction terms of the passed fraction.
func recordTerms(radius uint, f fraction, cache *[CACHESIZE]uint64, data *map[uint]uint64) {
	contFrac := getContinuedFrac(f.a, f.b)
	fracAmt := uint64(math.Sqrt(float64((radius * radius)) / float64((f.a*f.a + f.b*f.b))))

	for _, n := range contFrac {
		if n < CACHESIZE {
			cache[n] += fracAmt * 2
		} else {
			(*data)[n] += fracAmt * 2 //считаются элементы от a/b и b/a
		}
	}
}

// flushCache records weights from cache to data, cache is then emptied.
func (p *program) flushCache(cache *[CACHESIZE]uint64) map[uint]uint64 {
	res := map[uint]uint64{}
	for key := uint(1); key < CACHESIZE; key++ {
		res[key] = cache[key]
		cache[key] = 0
	}
	return res
}

// gcd returns greatest common divisor of a and b.
func gcd(a, b uint) uint {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// getContinuedFrac returns the array which is holding terms of continued fraction for passed a and b, which represent fraction a/b.
func getContinuedFrac(a, b uint) []uint {
	var contFrac []uint = make([]uint, 0)
	for b > 0 {
		contFrac = append(contFrac, a/b)
		a, b = b, a%b
	}
	return contFrac
}
