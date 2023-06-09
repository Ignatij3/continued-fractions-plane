package main_test

import (
	"testing"
)

const thousand = 1000

func run(size int, gcd func(int, int) int, b *testing.B) {
	for a := 1; a <= size; a++ {
		for b := 1; b <= size; b++ {
			gcd(a, b)
		}
	}
}

func benchmark(gcd func(int, int) int, b *testing.B) {
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
				run(test.size, gcd, b)
			}
		})
	}
}
