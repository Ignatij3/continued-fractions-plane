package plane

import (
	"math"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// sets cache for numbers [0; CACHESIZE-1].
const CACHESIZE = 10 + 1

// Start initializes calculations with given concurrent worker amount.
// Start does not block. To know, when have calculations finished, use NotifyOnFinish.
// If calculations are already running, Start does nothing.
// The function also calls interrupt listener, which will intercept one keyboard interrupt to correctly stop calculations.
func (p *Plane) Start() {
	if p.running {
		return
	}
	p.running = true

	// to account for diagonal
	if p.weights[1] == 0 {
		p.weights[1] += uint64(float64(p.pcfg.Radius) / math.Sqrt2)
	}

	reschan := p.initializeWorkers()
	go p.processResults(reschan)
	go p.interruptListener()
}

// interruptListener intercepts os.Interrupt and SIGTERM to correctly stop calculations.
// If calculations are stopped by other means, it exits.
func (p *Plane) interruptListener() {
	termination := make(chan os.Signal, 1)
	signal.Notify(termination, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-termination:
			p.Stop()
			return
		case <-p.processingFinish:
			return
		case <-p.exit:
			return
		}
	}
}

// Stop prematurely stops calculations.
// When Stop exits, it is safe to read obtained data.
func (p *Plane) Stop() {
	if !p.running {
		return
	}

	close(p.exit)
	p.running = false
	<-p.processingFinish
}

// processResults receives continued fraction element weights through reschan and passes them to underlying storage.
// On exit, processResults updates state if needed and sends signal that calculations have finished.
func (p *Plane) processResults(reschan chan map[uint]uint64) {
	cacheAmt := int(p.pcfg.Workers/10) + 1
	cacheDrain := make(chan map[uint]uint64, cacheAmt)

	cachesync := &sync.WaitGroup{}
	for i := 0; i < cacheAmt; i++ {
		cachesync.Add(1)
		go func() {
			p.processToCache(reschan, cacheDrain)
			cachesync.Done()
		}()
	}

	go func() {
		cachesync.Wait()
		close(cacheDrain)
	}()

	for cachedata := range cacheDrain {
		for key, value := range cachedata {
			p.weights[key] += uint64(value)
		}
	}

	if p.IsFinished() {
		p.cleanup()
	} else {
		p.updateState()
	}
	p.running = false
	close(p.processingFinish)
}

// processToCache receives continued fraction elements through reschan.
// On exit, it flushes cache through cacheDrain.
func (p *Plane) processToCache(reschan, cacheDrain chan map[uint]uint64) {
	cache := make(map[uint]uint64)
	for res := range reschan {
		for key, value := range res {
			cache[key] += value
		}
	}
	cacheDrain <- cache
}

// initializeWorkers starts all workers and returns channel where obtained data would be sent.
func (p *Plane) initializeWorkers() chan map[uint]uint64 {
	reschan := make(chan map[uint]uint64, p.pcfg.Workers)
	jobs := p.getJobChan()

	linelock := &sync.Mutex{}
	workerAmountSync := &sync.WaitGroup{}
	for i := uint(0); i < p.pcfg.Workers; i++ {
		workerAmountSync.Add(1)
		go func() {
			p.compute(linelock, jobs, reschan)
			workerAmountSync.Done()
		}()
	}

	go func() {
		workerAmountSync.Wait()
		close(reschan)
	}()

	return reschan
}

// getJobChan creates and returns channel, through which job descriptions would be sent for workers.
func (p *Plane) getJobChan() chan uint {
	lineID := make(chan uint, p.pcfg.Workers)
	go func() {
		for id := p.pcfg.LastLine + 1; id <= p.pcfg.Radius; id++ {
			lineID <- id
		}
		close(lineID)
	}()
	return lineID
}

// compute counts continued fraction elements in area described by data passed through jobs channel. For performance reasons, only fractions above diagonal of square RxR are processed.
// It counts continued fraction elements of irreducible fractions number of times fractions are present in the first circle's quarter with radius R,
// it also takes into account that the reverse of a/b (where a > b) has the same continued fraction elements as a/b, but with the leading zero.
// Upon receiving exit signal, it finishes processing current line and exits.
func (p *Plane) compute(linelock *sync.Mutex, jobs chan uint, reschan chan map[uint]uint64) {
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

		case <-p.exit:
			break work

		default:
			processLine(p.pcfg.Radius, line, &cache, res)
			if line >= p.pcfg.LastLine {
				linelock.Lock()
				p.pcfg.LastLine = line
				linelock.Unlock()
			}
		}
	}

	reschan <- *res
	reschan <- p.flushCache(&cache)
}

// processLine counts continued fraction elements in the given horizontal line.
// Fractions are composed of whole points' coordinates: (x;y) => x/y.
func processLine(radius, y uint, cache *[CACHESIZE]uint64, res *map[uint]uint64) uint {
	rightBoundSquared := radius*radius - y*y
	if y <= uint(float64(radius)/math.Sqrt2) {
		rightBoundSquared = (y - 1) * (y - 1)
	}

	var x, xSqr uint
	for ; xSqr <= rightBoundSquared; x++ {
		if gcd(x, y) == 1 {
			recordElements(radius, fraction{a: y, b: x}, cache, res)
		}
		xSqr += x<<1 + 1
	}

	return x
}

// recordElements updates cache and map with data with information about continued fraction elements of the given fraction.
// Radius of circle is required to correctly calculate amount of fractions with the same continued fractions in the area.
func recordElements(radius uint, f fraction, cache *[CACHESIZE]uint64, data *map[uint]uint64) {
	contFrac := f.getContinuedFraction()
	fracAmt := uint64(math.Sqrt(float64((radius * radius)) / float64((f.a*f.a + f.b*f.b))))

	for _, n := range contFrac {
		if n < CACHESIZE {
			cache[n] += fracAmt * 2
		} else {
			(*data)[n] += fracAmt * 2 //считаются элементы от a/b и b/a
		}
	}
}

// flushCache flushes weights from cache to data map and returns it.
func (p *Plane) flushCache(cache *[CACHESIZE]uint64) map[uint]uint64 {
	res := map[uint]uint64{}
	for key := uint(1); key < CACHESIZE; key++ {
		res[key] = cache[key]
		cache[key] = 0
	}
	return res
}
