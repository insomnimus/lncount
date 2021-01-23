package main

import (
	"flag"
	"fmt"
	globber "github.com/mattn/go-zglob"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
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

var liner = regexp.MustCompile(`[\n|\r]`)

func countLines(name string) int {
	data, err := ioutil.ReadFile(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening file: %s\n", err)
		os.Exit(1)
	}
	return len(liner.FindAllIndex(data, -1))
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
	numberJobs := len(files)
	jobs := make(chan string, numberJobs)
	results := make(chan int, numberJobs)
	workerN := 10
	if workerN > len(files) {
		workerN = len(files)
	}
	for i := 0; i <= workerN; i++ {
		go worker(jobs, results)
	}
	for _, j := range files {
		jobs <- j
	}
	close(jobs)
	for a := 1; a <= numberJobs; a++ {
		count += <-results
	}

	msg := fmt.Sprintf("%d lines in %d file", count, len(files))
	if len(files) > 1 {
		msg += "s"
	}
	fmt.Println(msg)
}

func worker(jobs <-chan string, results chan<- int) {
	for j := range jobs {
		results <- countLines(j)
	}
}
