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

// sets cache for numbers [0; CACHESIZE-1].
const CACHESIZE = 10 + 1

var (
	rectlock      sync.Mutex
	processingEnd *sync.WaitGroup = &sync.WaitGroup{}
)

// borders contains information about rectangle, defined by four coordinates.
type borders struct {
	XLeft, XRight uint
	YLow, YUpper  uint
}

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
	go func() {
		logger.Println("INFO: Setting up state backup")
		defer close(finish)

		termination := make(chan os.Signal, 1)
		signal.Notify(termination, os.Interrupt, syscall.SIGTERM)

		for {
			select {
			case <-time.After(30 * time.Minute):
				logger.Printf("INFO: State after 30 minutes:\ncells: %d\nworkers: %d\nN: %d\nlastRect: %v\nlastDiag: %v\n", p.CELLS, p.WORKERS, p.N, p.LastRect, p.LastDiag)
			case term := <-termination:
				logger.Printf("INFO: Process has been interrupted with %v, cleaning up\n", term)
				triggerExit()
				processingEnd.Wait()
				p.updateState()
				os.Exit(0)
			case <-finish:
				return
			}
		}
	}()

	diagonal := make(chan borders, p.CELLS)
	rects := make(chan borders, p.WORKERS)

	reschan := p.distributeJobs(diagonal, rects, exit, triggerExit)
	p.processResults(reschan)

	finish <- struct{}{}
	<-finish
}

// processResults receives term weights through reschan and adds them to underlying array.
func (p *program) processResults(reschan chan map[uint]uint32) {
	processingEnd.Add(1)
	defer processingEnd.Done()

	cachesync := &sync.WaitGroup{}
	cachesync.Add(int(p.CELLS))
	cacheDrain := make(chan map[uint]uint32, p.CELLS)

	go func() {
		cachesync.Wait()
		close(cacheDrain)
	}()

	for i := uint(0); i < p.CELLS; i++ {
		go p.processToCache(reschan, cacheDrain, cachesync)
	}

	for res := range reschan {
		for key, value := range res {
			p.weights[key] += uint64(value)
		}
	}

	for cachedata := range cacheDrain {
		for key, value := range cachedata {
			p.weights[key] += uint64(value)
		}
	}
}

// processToCache receives results through reschan and sends them back through cacheDrain on completion.
func (p *program) processToCache(reschan, cacheDrain chan map[uint]uint32, cachesync *sync.WaitGroup) {
	defer cachesync.Done()
	cache := make(map[uint]uint32)
	for res := range reschan {
		for key, value := range res {
			cache[key] += value
		}
	}
	cacheDrain <- cache
}

// initJobs passes rectangles that need to be processed through diagonal and rects channels.
// Rects that share diagonal with N×N plane are sent to channel diagonal,
// any other rects that lie strictly above diagonal are sent to channel rects.
func (p *program) initJobs(diagonal, rects chan borders) {
	logger.Println("INFO: Initializing bounds")
	if p.N <= 1000 {
		p.CELLS = 1
		diagonal <- borders{
			XLeft:  1,
			YLow:   1,
			XRight: p.N,
			YUpper: p.N,
		}
		close(diagonal)
		close(rects)
		return
	}

	if p.N/p.CELLS < 1 {
		p.CELLS = p.N / 10
	}
	step := uint(math.Ceil(float64(p.N) / float64(p.CELLS)))

	go p.initDiagonal(diagonal, step)
	go p.initAboveDiagonal(rects, step)
}

// initDiagonal sends squares which need to be processed in passed channel. All squares share diagonal with N×N plane.
func (p *program) initDiagonal(diagonal chan borders, step uint) {
	var a uint = p.LastDiag + 1
	for ; a < p.N-step; a += step + 1 {
		diagonal <- borders{
			XLeft:  a,
			XRight: a + step,
			YLow:   a,
			YUpper: a + step,
		}
		if !NDEBUG {
			logger.Printf("DEBUG: diagonal: %v\n", borders{XLeft: a, XRight: a + step, YLow: a, YUpper: a + step})
		}
	}

	if a < p.N {
		diagonal <- borders{
			XLeft:  a,
			XRight: p.N,
			YLow:   a,
			YUpper: p.N,
		}
		if !NDEBUG {
			logger.Printf("DEBUG: diagonal: %v\n", borders{XLeft: a, XRight: p.N, YLow: a, YUpper: p.N})
		}
	}

	close(diagonal)
}

// initAboveDiagonal sends rectangles which need to be processed in passed channel. All rectangles lie strictly above diagonal.
func (p *program) initAboveDiagonal(rects chan borders, step uint) {
	var x, y, yset uint = 1, 1, p.LastRect.YUpper + 1

	if p.LastRect.XLeft == 0 {
		p.LastRect.XLeft = 1
	}
	if p.LastRect.YUpper == 0 {
		yset = x + step + 1
	}

	for x = p.LastRect.XLeft; x < p.N-(step+1); x += step + 1 {
		for y = yset; y < p.N-(step+1); y += step + 1 {
			rects <- borders{
				XLeft:  x,
				XRight: x + step,
				YLow:   y,
				YUpper: y + step,
			}
			if !NDEBUG {
				logger.Printf("DEBUG: rects: %v\n", borders{XLeft: x, XRight: x + step, YLow: y, YUpper: y + step})
			}
		}
		rects <- borders{
			XLeft:  x,
			XRight: x + step,
			YLow:   y,
			YUpper: p.N,
		}
		if !NDEBUG {
			logger.Printf("DEBUG: rects: %v\n", borders{XLeft: x, XRight: x + step, YLow: y, YUpper: p.N})
		}
		yset = x + 2*(step+1)
	}
	close(rects)
}

// distributeJobs distributes rectangles among workers and returns channel where the results would be sent.
func (p *program) distributeJobs(diagonal, rects chan borders, exit context.Context, triggerExit context.CancelFunc) chan map[uint]uint32 {
	logger.Println("INFO: Initializing and distributing jobs")

	if p.weights == nil {
		p.weights = make([]uint64, p.N+1)
		p.weights[1] = uint64(p.N)
	}

	reschan := make(chan map[uint]uint32, p.WORKERS/10)

	p.initJobs(diagonal, rects)

	ressync := &sync.WaitGroup{}
	ressync.Add(int(p.WORKERS))

	go p.compute(ressync, diagonal, reschan, exit, p.onDiagonal)
	for i := uint(1); i < p.WORKERS; i++ {
		go p.compute(ressync, rects, reschan, exit, p.aboveDiagonal)
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

// computeLines counts continued fraction terms in rectangles above the diagonal of the plane N×N.
// It counts terms of irreducible fractions number of times fractions are present in the plane N×N, it also takes into account that the reverse of a/b (where a > b)
// has the same terms as a/b, but with the leading zero.
func (p *program) compute(ressync *sync.WaitGroup, jobs chan borders, reschan chan map[uint]uint32, exit context.Context, traverseRect func(borders, *[CACHESIZE]uint32, *map[uint]uint32)) {
	defer ressync.Done()
	var cache [CACHESIZE]uint32

work:
	for rect := range jobs {
		select {
		case <-exit.Done():
			break work
		default:
			res := &map[uint]uint32{}
			traverseRect(rect, &cache, res)
			reschan <- *res
		}
	}

	reschan <- p.flushCache(&cache)
}

// aboveDiagonal counts continued fraction terms above the diagonal of the plane N×N, the jobs's squares must be strictly above the diagonal (not touching it) of the plane.
func (p *program) aboveDiagonal(rect borders, cache *[CACHESIZE]uint32, res *map[uint]uint32) {
	for b := rect.XLeft; b <= rect.XRight; b++ {
		for a := rect.YLow; a <= rect.YUpper; a++ {
			if gcd(a, b) == 1 {
				recordTerms(p.N, fraction{a: a, b: b}, cache, res)
			}
		}
	}

	if rect.XLeft > p.LastRect.XLeft || (rect.XLeft == p.LastRect.XLeft && rect.YUpper > p.LastRect.YUpper) {
		rectlock.Lock()
		p.LastRect = rect
		rectlock.Unlock()
	}
}

// onDiagonal counts continued fraction terms right above the diagonal (excluding it) of the plane N×N, the jobs's squares must share the diagonal with that of the plane.
func (p *program) onDiagonal(rect borders, cache *[CACHESIZE]uint32, res *map[uint]uint32) {
	for b := rect.XLeft; b <= rect.XRight; b++ {
		for a := b + 1; a <= rect.YUpper; a++ {
			if gcd(a, b) == 1 {
				recordTerms(p.N, fraction{a: a, b: b}, cache, res)
			}
		}
	}
	p.LastDiag = rect.XRight
}

// recordTerms updates cache and data map with information about continued fraction terms of the passed fraction.
func recordTerms(N uint, f fraction, cache *[CACHESIZE]uint32, data *map[uint]uint32) {
	contFrac := getContinuedFrac(f.a, f.b)
	fracAmt := uint32(N / f.a)

	for _, n := range contFrac {
		if n < CACHESIZE {
			cache[n] += fracAmt * 2
		} else {
			(*data)[n] += fracAmt * 2 //считаются элементы от a/b и b/a
		}
	}
}

// flushCache records weights from cache to data, cache is then emptied.
func (p *program) flushCache(cache *[CACHESIZE]uint32) map[uint]uint32 {
	res := map[uint]uint32{}
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
