package main_test

import (
	"testing"
)

const thousand = 1000

func run(size uint64, root func(uint64) uint64, b *testing.B) {
	for a := uint64(1); a <= size; a++ {
		for b := uint64(1); b <= size; b++ {
			root(size*size - a*a)
		}
	}
}

func benchmark(root func(uint64) uint64, b *testing.B) {
	var testset = []struct {
		name string
		size uint64
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
				run(test.size, root, b)
			}
		})
	}
}
