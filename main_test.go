package main

import (
	"os"
	"testing"
)

func Benchmark_main(b *testing.B) {
	home, err := os.UserHomeDir()
	if err != nil {
		b.Fatal(err)
	}

	os.Args = []string{"dups", home}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		main()
	}
}

// func BenchmarkExperiments(b *testing.B) {
// 	const bufsize = 200

// 	for i := 0; i < b.N; i++ {
// 		buf := make([]byte, bufsize)
// 		m := map[uint64]int{}
// 		sfmt := "%016x%0" + fmt.Sprint(bufsize*2) + "x"

// 		filepath.WalkDir(folder, func(path string, d fs.DirEntry, err error) error {
// 			if !d.Type().IsRegular() {
// 				return nil
// 			}

// 			f, err := os.Open(path)
// 			if err != nil {
// 				return nil
// 			}

// 			s, err := f.Stat()
// 			if err != nil {
// 				return nil
// 			}

// 			_, err = f.Read(buf)
// 			if err != nil {
// 				return nil
// 			}

// 			k := crc64.Checksum([]byte(fmt.Sprintf(sfmt, s.Size(), buf)), crc64.MakeTable(crc64.ISO))
// 			fmt.Printf("%016x\n", k)
// 			_, ok := m[k]
// 			if !ok {
// 				m[k] = 0
// 			}
// 			m[k]++

// 			return nil
// 		})

// 		t := 0
// 		for _, v := range m {
// 			if v > 1 {
// 				t += v
// 			}
// 		}

// 		fmt.Printf("\nfound %d potential duplicates\n\n", t)
// 	}
// }

// func Benchmark_Walk(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		filepath.Walk(folder, func(path string, info fs.FileInfo, err error) error {
// 			fmt.Println(path)
// 			return nil
// 		})
// 	}
// }

// func Benchmark_WalkDir(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		filepath.WalkDir(folder, func(path string, d fs.DirEntry, err error) error {
// 			fmt.Println(path)
// 			return nil
// 		})
// 	}
// }

// func Benchmark_WalkDir_GetInfoForRegularFiles(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		counts := map[int64]int{}

// 		filepath.WalkDir(folder, func(path string, d fs.DirEntry, err error) error {
// 			fmt.Println(path)
// 			if filepath.Base(path) == ".git" {
// 				return filepath.SkipDir
// 			}

// 			if !d.Type().IsRegular() {
// 				return nil
// 			}

// 			s, err := os.Stat(path)
// 			if err != nil {
// 				return nil
// 			}

// 			counts[s.Size()]++

// 			_ = s
// 			return nil
// 		})

// 		uniques := 0
// 		duplicates := 0
// 		total := 0

// 		fmt.Println()
// 		fmt.Println()

// 		for size, count := range counts {
// 			if count > 1 {
// 				fmt.Printf("size: %d - count: %d\n", count, size)
// 				duplicates += count
// 			} else {
// 				uniques++
// 			}
// 			total += count
// 		}

// 		fmt.Printf("unique size: %d, same size: %d, total: %d\n", uniques, duplicates, total)
// 	}
// }

// func Benchmark_WalkDir_Read10B(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		buf := make([]byte, 10)

// 		filepath.WalkDir(folder, func(path string, d fs.DirEntry, err error) error {
// 			fmt.Println(path)

// 			if !d.Type().IsRegular() {
// 				return nil
// 			}

// 			f, err := os.Open(path)
// 			if err != nil {
// 				return nil
// 			}

// 			f.Read(buf)

// 			return nil
// 		})
// 	}
// }
