package main

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

func main() {
	pathsBySize := map[int64][]string{}

	err := filepath.WalkDir(os.Args[1], func(path string, d fs.DirEntry, err error) error {
		if !d.Type().IsRegular() {
			return nil
		}

		s, err := os.Stat(path)
		if err != nil {
			log.Print(err)
		}

		if pathsBySize[s.Size()] == nil {
			pathsBySize[s.Size()] = []string{}
		}

		pathsBySize[s.Size()] = append(pathsBySize[s.Size()], path)
		return nil
	})

	if err != nil {
		log.Print(err)
		return
	}

	for size, paths := range pathsBySize {
		if len(paths) < 2 {
			continue
		}

		if size == 0 {
			continue
		}

		for i := 0; i < len(paths)-1; i++ {
			dups := []string{}
			p1 := paths[i]

			for j := i + 1; j < len(paths); j++ {
				p2 := paths[j]
				func() {
					f1, err := os.Open(p1)
					if err != nil {
						log.Print(err)
						return
					}
					defer f1.Close()

					f2, err := os.Open(p2)
					if err != nil {
						log.Print(err)
						return
					}
					defer f2.Close()

					eq, err := equalReaders(f1, f2)
					if err != nil {
						log.Print(err)
						return
					}

					if eq {
						dups = append(dups, p2)
					}
				}()
			}

			if len(dups) > 0 {
				fmt.Println("these files are equal:")
				fmt.Println(" ", p1)
				for _, dup := range dups {
					fmt.Println(" ", dup)
				}
				fmt.Println()
			}

		}
	}
}

func equalReaders(r1 io.Reader, r2 io.Reader) (bool, error) {
	buf1 := make([]byte, 4096)
	buf2 := make([]byte, 4096)

	for {
		n1, err1 := r1.Read(buf1)
		if err1 != nil && err1 != io.EOF {
			return false, err1
		}

		n2, err2 := r2.Read(buf2)
		if err2 != nil && err2 != io.EOF {
			return false, err2
		}

		if err1 == io.EOF {
			break
		}

		if n1 != n2 {
			return false, nil
		}

		if !bytes.Equal(buf1[:n1], buf2[:n1]) {
			return false, nil
		}
	}

	return true, nil
}
