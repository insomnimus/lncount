package main

import (
	"flag"
	"fmt"
	"github.com/gobwas/glob"
	"github.com/loveleshsharma/gohive"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var files []string
var count int
var (
	rec bool
	g   glob.Glob
)

func collectFiles() {
	if !rec {
		fs, err := ioutil.ReadDir("./")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		for _, f := range fs {
			if !f.IsDir() {
				if g.Match(f.Name()) {
					files = append(files, f.Name())
				}
			}
		}
		return
	}

	err := filepath.Walk("./", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if g.Match(info.Name()) {
				files = append(files, path)
			}
		}
		return nil
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var mux sync.Mutex

func countLines(name string) {
	data, err := ioutil.ReadFile(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening file: %s\n", err)
		os.Exit(1)
	}
	cnt := strings.Count(string(data), "\n")
	if cnt == 0 {
		cnt += strings.Count(string(data), "\r\n")
	}
	if cnt == 0 {
		cnt += strings.Count(string(data), "\r")
	}
	mux.Lock()
	count += cnt
	mux.Unlock()
}

func main() {
	flag.BoolVar(&rec, "r", false, "recursively search all files and subfolders")
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "missing argument: pattern")
		return
	}
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "too many arguments")
		return
	}
	pattern := args[0]
	g = glob.MustCompile(pattern)
	collectFiles()
	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "no files found matching %s\n", pattern)
		return
	}
	hive := gohive.NewFixedSizePool(5)
	var wg sync.WaitGroup
	for _, f := range files {
		exe := func() {
			defer wg.Done()
			countLines(f)
		}
		wg.Add(1)
		hive.Submit(exe)
	}
	wg.Wait()
	fmt.Printf("%d lines\n", count)
}
