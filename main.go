package main

import (
	"flag"
	"fmt"
	"github.com/loveleshsharma/gohive"
	globber "github.com/mattn/go-zglob"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sync"
)

var files []string
var count int
var (
	exeName     string = "lncount"
	filePattern string
)

func collectFiles() {
	var err error
	files, err = globber.Glob(filePattern)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

var mux sync.Mutex

var liner = regexp.MustCompile(`[\n|\r]`)

func countLines(name string) {
	data, err := ioutil.ReadFile(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening file: %s\n", err)
		os.Exit(1)
	}
	matches := liner.FindAllIndex(data, -1)
	mux.Lock()
	count += len(matches)
	mux.Unlock()
}

func helpMsg() {
	fmt.Fprintf(os.Stderr, "usage: %s <filename pattern>\n"+
		"supports glob patterns\n",
		exeName)
}

func main() {
	help := flag.Bool("h", false, "display help")
	help2 := flag.Bool("help", false, "display help")
	flag.Parse()
	exeName = filepath.Base(os.Args[0])
	if *help || *help2 {
		helpMsg()
		return
	}
	args := flag.Args()
	if len(args) == 0 {
		helpMsg()
		return
	}
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "too many arguments")
		return
	}
	filePattern = args[0]
	collectFiles()
	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "no files found matching %s\n", filePattern)
		return
	}
	hive := gohive.NewFixedSizePool(10)
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
	msg:= fmt.Sprintf("%d lines in %d file", count, len(files))
	if len(files)> 1{
		msg+= "s"
	}
	fmt.Println(msg)
	
}
