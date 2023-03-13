package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"sync"
)

const cpuProf = false

func main() {
	if cpuProf {
		f, err := os.Create("cpu.prof")
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	pathBySize := map[int64][]string{}

	err := filepath.WalkDir(os.Args[1], func(path string, d fs.DirEntry, err error) error {
		if !d.Type().IsRegular() {
			return nil
		}

		s, err := os.Stat(path)
		if err != nil {
			log.Print(err)
			return nil
		}

		if pathBySize[s.Size()] == nil {
			pathBySize[s.Size()] = []string{}
		}
		pathBySize[s.Size()] = append(pathBySize[s.Size()], path)

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	for size := range pathBySize {
		mu := sync.Mutex{}
		hashes := map[string][]string{}
		limit := make(chan struct{}, runtime.NumCPU())
		wg := sync.WaitGroup{}

		for _, path := range pathBySize[size] {
			path := path
			wg.Add(1)
			limit <- struct{}{}

			go func() {
				defer wg.Done()
				defer func() { <-limit }()

				f, err := os.Open(path)
				if err != nil {
					log.Print(err)
					return
				}
				defer f.Close()

				h := sha1.New()

				_, err = io.Copy(h, f)
				if err != nil {
					log.Print(err)
					return
				}

				sum := h.Sum(nil)
				key := hex.EncodeToString(sum)

				func() {
					mu.Lock()
					defer mu.Unlock()

					if hashes[key] == nil {
						hashes[key] = []string{}
					}

					hashes[key] = append(hashes[key], path)
				}()
			}()
		}
		wg.Wait()

		for key, paths := range hashes {
			if len(paths) <= 1 {
				continue
			}

			fmt.Printf("These files have the same hash and size: %s %d\n", key, size)
			for _, path := range paths {
				fmt.Println(" ", path)
			}
			fmt.Println()
		}
	}
}
