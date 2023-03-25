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
	wg := sync.WaitGroup{}

	_ = filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if !d.Type().IsRegular() {
			return nil
		}

		wg.Add(1)

		go func() {
			defer wg.Done()

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
	pathsBySum := map[uint64][]string{}

	buf := make([]byte, 200)

	for size, paths := range pathsBySize {
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

				pathsBySum[sum] = append(pathsBySum[sum], path)
			}()
		}
	}

	return pathsBySum
}

func finalHash(pathsBySum map[uint64][]string) {
	empties := []string{}

	for _, paths := range pathsBySum {
		if len(paths) == 1 {
			continue
		}

		pathsByHash := map[string][]string{}

		for _, path := range paths {
			func() {
				f, err := os.Open(path)
				if err != nil {
					log.Printf("failed to open %s: %v", path, err)
					return
				}
				defer f.Close()

				h := sha1.New()
				_, err = io.Copy(h, f)
				if err == io.EOF {
					empties = append(empties, path)
				} else if err != nil {
					log.Printf("failed to read %s: %v", path, err)
					return
				}

				sum := hex.EncodeToString(h.Sum(nil))
				pathsByHash[sum] = append(pathsByHash[sum], path)
			}()
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
		}
	}
}
