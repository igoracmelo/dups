package main

import (
	"os"
	"testing"
)

func Benchmark_main(b *testing.B) {
	os.Args = []string{"dups", "./samples"}

	for i := 0; i < b.N; i++ {
		main()
	}
}
