package main

import (
	"os/exec"
	"testing"
)

func BenchmarkDups_Defaults(b *testing.B) {
	err := exec.Command("go", "build", "-o", "dups", ".").Run()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err = exec.Command("./dups", "./samples").Run()
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkDups_Fast(b *testing.B) {
	err := exec.Command("go", "build", "-o", "dups", ".").Run()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err = exec.Command("./dups", "--fast", "./samples").Run()
		if err != nil {
			b.Error(err)
		}
	}
}
