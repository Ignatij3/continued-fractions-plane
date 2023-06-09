package main_test

import (
	"testing"
)

func runVanilla(size int) map[int]int {
	var (
		res map[int]int = make(map[int]int)
		f   fraction
	)

	for f.b = 1; f.b <= size; f.b++ {
		for f.a = 1; f.a <= size; f.a++ {
			contFrac := getContinuedFrac(f)
			for _, n := range contFrac {
				res[n]++
			}
		}
	}

	return res
}

func BenchmarkVanilla(b *testing.B) {
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

	for _, test := range testset {
		b.Run(test.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				runVanilla(test.size)
			}
		})
	}
}
