package plane

import "strconv"

// Plane represents a set of all fractions which would be processed.
// It is defined as first quarter of circle with radius Radius.
type Plane struct {
	pcfg             planeConfig
	exit             chan struct{}
	processingFinish chan struct{}

	running bool
	weights []uint64
}

// planeConfig contains data required to distinguish different planes and different states of one plane.
type planeConfig struct {
	LastLine uint
	Workers  uint
	Radius   uint
	rstr     string
}

// InitNewPlane creates and returns new plane with passed configuration.
func InitNewPlane(radius, workers uint) *Plane {
	return &Plane{
		pcfg: planeConfig{
			LastLine: 0,
			Workers:  workers,
			Radius:   radius,
			rstr:     strconv.Itoa(int(radius)),
		},
		exit:             make(chan struct{}),
		processingFinish: make(chan struct{}),
		running:          false,
		weights:          make([]uint64, radius+1),
	}
}

// NotifyOnFinish returns channel which would receive signal when all calculations have finished.
// After receiving signal, it is safe to read obtained data.
func (p Plane) NotifyOnFinish() <-chan struct{} {
	return p.processingFinish
}

// GetResults returns slice with total occurrences of elements of continued fractions in set plane.
// If calculations are running, the function returns nil.
func (p Plane) GetResults() []uint64 {
	if p.running {
		return nil
	}

	wghtCopy := make([]uint64, len(p.weights))
	copy(wghtCopy, p.weights)
	return wghtCopy
}

// IsRunning reports whether calculations are still running.
func (p Plane) IsRunning() bool {
	return p.running
}

// IsFinished reports whether the module has processed whole plane.
func (p Plane) IsFinished() bool {
	return p.pcfg.LastLine == p.pcfg.Radius
}
