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
	"runtime/pprof"
	"sync"
)

var fast bool
var minSize int64

func main() {

	log.SetPrefix("")
	log.SetFlags(0)

	flag.BoolVar(&fast, "fast", false, "Process faster, but uses a less trustworthy algorithm")
	flag.Int64Var(&minSize, "size", 0, "Will only process files greater than this value")
	flag.Parse()

	var dir string
	var err error

	f, err := os.Create("cpu.prof")
	if err != nil {
		log.Panic(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

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

	runner := newRunner(dir, fast, minSize)

	go runner.getFileInfos()
	go runner.hashFiles()

	runner.processResults()
	runner.printResults()
}

type runner struct {
	newHash     func() hash.Hash
	hashSize    int
	dirpath     string
	showEmpties bool
	minSize     int64
	fast        bool
	dups        map[string][]dup
	empties     []string
	jobs        chan hashJob
	results     chan hashResult
}

type dup struct {
	size int64
	path string
}

type hashJob struct {
	path string
	size int64
}

type hashResult struct {
	key  string
	path string
	size int64
}

func newRunner(dirpath string, fast bool, minSize int64) *runner {
	r := &runner{
		dirpath:     dirpath,
		fast:        fast,
		minSize:     minSize,
		showEmpties: false, // TODO
		empties:     []string{},
		dups:        map[string][]dup{},
		jobs:        make(chan hashJob, 1),
		results:     make(chan hashResult, 1),
	}

	if fast {
		r.newHash = func() hash.Hash { return crc32.NewIEEE() }
		r.hashSize = crc32.Size
	} else {
		r.newHash = sha1.New
		r.hashSize = sha1.Size
	}

	return r
}

func (r *runner) getFileInfos() error {
	wg := sync.WaitGroup{}

	err := filepath.WalkDir(r.dirpath, func(path string, d fs.DirEntry, err error) error {
		if !d.Type().IsRegular() {
			return nil
		}

		wg.Add(1)

		go func() {
			defer wg.Done()

			s, err := os.Stat(path)
			if err != nil {
				log.Printf("failed to get stat: %s\n%v", s.Name(), err)
				return
			}

			// log.Printf(">>> job: %s, %d\n", path, s.Size())
			r.jobs <- hashJob{
				path: path,
				size: s.Size(),
			}
		}()

		return nil
	})

	wg.Wait()
	close(r.jobs)

	return err
}

func (r *runner) hashFiles() {
	limit := make(chan struct{}, 1)
	wg := sync.WaitGroup{}

	for j := range r.jobs {
		j := j
		wg.Add(1)

		go func() {
			limit <- struct{}{}
			defer func() { <-limit }()
			defer wg.Done()

			h := r.newHash()
			f, err := os.Open(j.path)
			if err != nil {
				log.Printf("failed to open %s: %v", j.path, err)
				return
			}

			log.Printf(">>> hashing %s\n", j.path)

			_, err = io.Copy(h, f)
			if err != nil {
				log.Printf("failed to read %s: %v", j.path, err)
			}

			sum := h.Sum(nil)
			key := hex.EncodeToString(sum) + fmt.Sprintf("%016x", j.size)

			r.results <- hashResult{
				key:  key,
				path: j.path,
				size: j.size,
			}
		}()
	}

	wg.Wait()
	close(r.results)
}

func (r *runner) processResults() {
	results := make(chan hashResult, 1)

	for res := range results {
		if r.dups[res.key] == nil {
			r.dups[res.key] = []dup{}
		}
		this := dup{
			size: res.size,
			path: res.path,
		}
		r.dups[res.key] = append(r.dups[res.key], this)
	}
}

func (r *runner) printResults() {
	for k, v := range r.dups {
		if len(v) == 1 {
			continue
		}

		fmt.Printf("these files have the same hash (%s):\n", k[:r.hashSize*2])
		for _, p := range v {
			fmt.Printf("  %d  %s\n", p.size, p.path)
		}
		fmt.Println()
	}

	if r.showEmpties && len(r.empties) > 0 {
		fmt.Println("these files are empty:")
		for _, v := range r.empties {
			fmt.Println(" ", v)
		}
	}
}
