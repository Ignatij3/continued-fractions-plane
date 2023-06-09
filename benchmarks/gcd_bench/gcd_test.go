package main_test

import (
	"testing"
)

func gcdLoop(a, b int) int {
	for a != b {
		if a > b {
			a -= b
		} else {
			b -= a
		}
	}

	return a
}

func gcdMod(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func gcdRec(a, b int) int {
	if a == 0 {
		return b
	}
	return gcdRec(b%a, a)
}

func gcdExt(a, b int) int {
	var quot int

	oldr, r := a, b
	olds, s := 1, 0
	oldt, t := 0, 1

	for r != 0 {
		quot = oldr / r
		oldr, r = r, oldr-(quot*r)
		olds, s = s, olds-(quot*s)
		oldt, t = t, oldt-(quot*t)
	}

	return oldr
}

func BenchmarkMod(b *testing.B) {
	benchmark(gcdMod, b)
}

func BenchmarkExt(b *testing.B) {
	benchmark(gcdRec, b)
}

func BenchmarkRec(b *testing.B) {
	benchmark(gcdExt, b)
}

func BenchmarkLoop(b *testing.B) {
	benchmark(gcdLoop, b)
}
