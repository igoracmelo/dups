package main

import (
	"crypto/sha1"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

func main() {
	log.SetPrefix("")
	log.SetFlags(0)

	flag.Parse()

	var dir string
	var err error

	if flag.NArg() != 1 {
		dir, err = os.Getwd()
		if err != nil {
			log.Panic(err)
		}
	} else {
		dir = flag.Arg(0)
	}

	s, err := os.Stat(dir)
	if err != nil {
		log.Panic(err)
	}

	if !s.IsDir() {
		log.Panicf("not a directory: %s", s.Name())
	}

	type dup struct {
		path string
		size int64
	}

	dups := map[string][]dup{}
	empties := []string{}

	showEmpties := false

	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if !d.Type().IsRegular() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			log.Printf("failed to get info from file: %s", path)
		}

		if showEmpties && info.Size() == 0 {
			empties = append(empties, path)
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			log.Printf("failed to open %s: %v", path, err)
			return nil
		}
		defer f.Close()

		h := sha1.New()
		_, err = io.Copy(h, f)
		if err != nil {
			log.Printf("failed to read file %s: %v", path, err)
		}

		sum := h.Sum(nil)
		x := hex.EncodeToString(sum)

		this := dup{
			path: path,
			size: info.Size(),
		}

		if dups[x] != nil {
			other := dups[x][0]
			if this.size != other.size {
				fmt.Println()
				log.Printf("size: %d, path: %s, hash: %s\n", other.size, other.path, x)
				log.Printf("size: %d, path: %s, hash: %s\n", this.size, this.path, x)
				log.Panicln("found hash collision??")
			}
			dups[x] = append(dups[x], this)

		} else {
			dups[x] = []dup{this}
		}

		return nil
	})

	for k, v := range dups {
		if len(v) == 1 {
			continue
		}

		fmt.Printf("these files have the same hash (%s):\n", k)
		for _, p := range v {
			fmt.Printf("  %d  %s\n", p.size, p.path)
		}
		fmt.Println()
	}

	if showEmpties && len(empties) > 0 {
		fmt.Println("these files are empty:")
		for _, v := range empties {
			fmt.Println(" ", v)
		}
	}
}
