package main

import (
	"bufio"
	"bytes"
	"fmt"
	globber "github.com/mattn/go-zglob"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var files []string
var count int
var (
	exeName     string = "lncount"
	filePattern string
)

func collectFiles() {
	// assume bash is the default shell for all
	// bash can expand the filenames itself so return
	if runtime.GOOS != "windows" {
		return
	}
	var err error
	files, err = globber.Glob(filePattern)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

func countLines(name string) int {
	f, err := os.Open(name)
	if err != nil {
		return 0
	}
	defer f.Close()
	buf := make([]byte, 32*1024)
	cnt := 0
	lineSep := []byte{'\n'}
	for {
		c, err := f.Read(buf)
		cnt += bytes.Count(buf[:c], lineSep)
		if err != nil {
			return cnt
		}
	}
}

func helpMsg() {
	fmt.Fprintf(os.Stderr, "usage: %s <filename pattern>\n"+
		"supports glob patterns\n"+
		"if piped, reads from stdin instead (cat main.go | lncount)\n",
		exeName)
}

func readStdin() {
	scanner := bufio.NewScanner(os.Stdin)
	cnt := 0
	for scanner.Scan() {
		cnt++
		if scanner.Err() != nil {
			fmt.Printf("%d lines\n", cnt)
			return
		}
	}
	fmt.Printf("%d lines\n", cnt)
}

func main() {
	// check if stdin is piped
	if fi, err := os.Stdin.Stat(); err == nil {
		if (fi.Mode() & os.ModeCharDevice) == 0 {
			readStdin()
			return
		}
	}

	exeName = filepath.Base(os.Args[0])
	if len(os.Args) == 1 {
		helpMsg()
		return
	}

LOOP:
	for _, arg := range os.Args[1:] {
		if arg == "-h" || arg == "--help" {
			helpMsg()
			return
		}
		switch runtime.GOOS {
		case "windows":
			if strings.Contains(arg, "*") {
				filePattern = arg
				continue LOOP
			}
			files = append(files, arg)
		default:
			files = append(files, arg)
		}
	}
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
