package main_test

import (
	"math"
	"testing"
)

func sqrtMath(x uint64) uint64 {
	return uint64(math.Sqrt(float64(x)))
}

func SqrtInverse(x uint64) uint64 {
	n := 1.0 / float64(x)

	n2, th := n*0.5, 1.5
	b := math.Float64bits(n)
	b = 0x5FE6EB50C7B537A9 - (b >> 1)
	f := math.Float64frombits(b)
	f *= th - (n2 * f * f)
	return uint64(f)
}

func SqrtFast(x uint64) (r uint64) {
	var b uint64
	//Fast way to make p highest power of 4 <= x
	p := x
	var n, v uint
	if p >= 1<<32 {
		v = uint(p >> 32)
		n = 32
	} else {
		v = uint(p)
	}

	if v >= 1<<16 {
		v >>= 16
		n += 16
	}
	if v >= 1<<8 {
		v >>= 8
		n += 8
	}
	if v >= 1<<4 {
		v >>= 4
		n += 4
	}
	if v >= 1<<2 {
		n += 2
	}
	p = 1 << n

	for ; p != 0; p >>= 2 {
		b = r | p
		r >>= 1
		if x >= b {
			x -= b
			r |= p
		}
	}
	return
}

const DELTA = 0.01
const INITIAL_Z = 50.0

func SqrtFast2(x uint64) uint64 {
	xf := float64(x)
	z := INITIAL_Z

	step := func() float64 {
		return z - (z*z-xf)/(2*z)
	}

	for zz := step(); math.Abs(zz-z) > DELTA; {
		z = zz
		zz = step()
	}
	return uint64(x)
}

func BenchmarkInverse(b *testing.B) {
	benchmark(SqrtInverse, b)
}

func BenchmarkFast(b *testing.B) {
	benchmark(SqrtFast, b)
}

func BenchmarkFast2(b *testing.B) {
	benchmark(SqrtFast2, b)
}

func BenchmarkMath(b *testing.B) {
	benchmark(sqrtMath, b)
}
