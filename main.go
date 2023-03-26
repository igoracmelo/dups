package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"hash/crc64"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

func main() {
	log.SetPrefix("")
	log.SetFlags(0)

	infos := make(chan fileInfo, 1)
	go getFileInfos(infos, os.Args[1])

	pathsBySize := map[int64][]string{}

	for info := range infos {
		pathsBySize[info.size] = append(pathsBySize[info.size], info.path)
	}

	pathsByHash := quickTriage(pathsBySize)
	finalHash(pathsByHash)
}

type fileInfo struct {
	path string
	size int64
}

func getFileInfos(output chan<- fileInfo, path string) {
	log.Println("INFO: finding files")
	defer log.Println("INFO: finding files done")

	wg := sync.WaitGroup{}
	limit := make(chan struct{}, 20)

	_ = filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if filepath.Base(path) == ".git" {
			return filepath.SkipDir
		}

		if !d.Type().IsRegular() {
			return nil
		}

		limit <- struct{}{}
		wg.Add(1)

		go func() {
			defer wg.Done()
			defer func() { <-limit }()

			s, err := os.Stat(path)
			if err != nil {
				log.Printf("failed to get stat for file %s: %v", path, err)
				return
			}

			output <- fileInfo{
				path: path,
				size: s.Size(),
			}
		}()

		return nil
	})

	wg.Wait()
	close(output)
}

func quickTriage(pathsBySize map[int64][]string) map[uint64][]string {
	log.Println("INFO: doing quick triage")
	defer log.Println("INFO: quick triage done")

	pathsBySum := map[uint64][]string{}

	buf := make([]byte, 200)
	type result struct {
		key uint64
		val string
	}

	results := make(chan result)
	limit := make(chan struct{}, 10)
	wg := sync.WaitGroup{}

	go func() {
		for size, paths := range pathsBySize {
			// log.Println("INFO: quick triage for files with size", size)
			size := size
			paths := paths

			limit <- struct{}{}
			wg.Add(1)

			go func() {
				defer func() { <-limit }()
				defer wg.Done()

				for _, path := range paths {
					func() {
						f, err := os.Open(path)
						if err != nil {
							log.Printf("failed to open file %s: %v", path, err)
							return
						}
						defer f.Close()

						n, err := f.Read(buf)
						if err == io.EOF {
							return
						}
						if err != nil {
							log.Printf("failed to read file %s: %v", path, err)
							return
						}

						sizeHex := fmt.Sprintf("%016x", size)
						sum := crc64.Checksum(append(buf[:n], []byte(sizeHex)...), crc64.MakeTable(crc64.ISO))
						// log.Println("INFO: sum computed", sum)

						results <- result{
							key: sum,
							val: path,
						}
					}()
				}
			}()
		}
		wg.Wait()
		close(results)
	}()

	for res := range results {
		pathsBySum[res.key] = append(pathsBySum[res.key], res.val)
	}

	return pathsBySum
}

func finalHash(pathsBySum map[uint64][]string) {
	// empties := []string{}
	log.Println("INFO: computing definitive hashes")
	t := 0

	for _, paths := range pathsBySum {
		if len(paths) == 1 {
			continue
		}

		pathsByHash := map[string][]string{}

		type result struct {
			key string
			val string
		}

		results := make(chan result)

		go func() {
			wg := sync.WaitGroup{}
			limit := make(chan struct{}, runtime.NumCPU())

			for _, path := range paths {
				wg.Add(1)
				limit <- struct{}{}
				path := path

				go func() {
					defer func() { <-limit }()
					defer wg.Done()

					f, err := os.Open(path)
					if err != nil {
						log.Printf("failed to open %s: %v", path, err)
						return
					}
					defer f.Close()

					h := sha1.New()
					_, err = io.Copy(h, f)
					if err == io.EOF {
						// empties = append(empties, path)
					} else if err != nil {
						log.Printf("failed to read %s: %v", path, err)
						return
					}

					sum := hex.EncodeToString(h.Sum(nil))
					results <- result{
						key: sum,
						val: path,
					}
				}()

			}

			wg.Wait()
			close(results)
		}()

		for res := range results {
			pathsByHash[res.key] = append(pathsByHash[res.key], res.val)
		}

		for _, paths := range pathsByHash {
			if len(paths) == 1 {
				continue
			}

			fmt.Println("These files have the same content:")
			for _, path := range paths {
				fmt.Println(" ", path)
			}
			fmt.Println()

			t += len(paths)
		}
	}

	fmt.Println("Total number of files with shared content:", t)
}
