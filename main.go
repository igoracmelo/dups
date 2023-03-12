package main

import (
	"crypto/sha1"
	"encoding/hex"
	"flag"
	"fmt"
	"hash"
	"hash/crc32"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

var fast bool
var size int64

func main() {

	log.SetPrefix("")
	log.SetFlags(0)

	flag.BoolVar(&fast, "fast", false, "Process faster, but uses a less trustworthy algorithm")
	flag.Int64Var(&size, "size", 0, "Will only process files greater than this value")
	flag.Parse()

	var dir string
	var err error
	var h hash.Hash
	var hSize int

	if fast {
		h = crc32.NewIEEE()
		hSize = crc32.Size
	} else {
		h = sha1.New()
		hSize = sha1.Size
	}

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	dir = flag.Arg(0)

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
		defer h.Reset()

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

		if info.Size() < size {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			log.Printf("failed to open %s: %v", path, err)
			return nil
		}
		defer f.Close()

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

		k := x + fmt.Sprintf("%016x", this.size)

		if dups[k] != nil {
			other := dups[k][0]
			if this.size != other.size {
				fmt.Println()
				log.Printf("size: %d, path: %s, hash: %s\n", other.size, other.path, x)
				log.Printf("size: %d, path: %s, hash: %s\n", this.size, this.path, x)
				log.Panicln("found hash collision??")
			}
			dups[k] = append(dups[k], this)

		} else {
			dups[k] = []dup{this}
		}

		return nil
	})

	for k, v := range dups {
		if len(v) == 1 {
			continue
		}

		fmt.Printf("these files have the same hash (%s):\n", k[:hSize*2])
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
