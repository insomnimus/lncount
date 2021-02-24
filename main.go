package main

import (
	"bufio"
	"bytes"
	"fmt"
	globber "github.com/mattn/go-zglob"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/insomnimus/lncount/engine"
)

var (
	excluder    = regexp.MustCompile(`^(\-e|\-\-exclude)=(.+)$`)
	files       []string
	count       int
	exeName     string = "lncount"
	filePattern string
	exclude     string
)

func collectFiles() {
	if exclude != "" {
		defer filterFiles()
	}
	// assume bash is the default shell for all
	// bash can expand the filenames itself so return
	if runtime.GOOS != "windows" {
		return
	}
	tempFiles, err := globber.Glob(filePattern)
	if err != nil && len(files) == 0 {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	files = append(files, tempFiles...)
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

func showHelp() {
	if runtime.GOOS == "windows" {
		if strings.HasSuffix(exeName, ".exe") {
			exeName = exeName[:len(exeName)-4]
		}
	}
	fmt.Fprintf(os.Stderr, `%s, counts lines
	
	usage: %s <files> [options]
	
	options are:
	-e, --exclude <pattern>: exclude files using basic regexp
	-h, --help: show this message
`, exeName, exeName)
	if runtime.GOOS != "windows" {
		fmt.Fprintln(os.Stderr, "note: the --exclude flags value should be quoted if your shell automatically expands glob patterns.\n"+
			"in contrast, when specifying files, do not quote the glob pattern, let the shell handle that one.")
	}
	os.Exit(0)
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

func worker(jobs <-chan string, results chan<- int) {
	for j := range jobs {
		results <- countLines(j)
	}
}

func filterFiles() {
	if len(files) == 0 {
		return
	}
	rex, err := engine.Compile(exclude, true)
	if err != nil {
		exit("%s", err)
	}
	var fs []string
	for _, f := range files {
		if rex.MatchString(f) {
			continue
		}
		fs = append(fs, f)
	}
	files = fs
}

func exit(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(2)
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
		showHelp()
	}
	args := os.Args[1:]

	for i := 0; i < len(args); i++ {
		a := args[i]
		switch a[0] {
		case '-':
			switch a {
			case "-h", "--help":
				showHelp()
			case "-e", "--exclude":
				if i+1 >= len(args) {
					exit("the --exclude flag is set but the value is missing")
				}
				exclude = args[i+1]
				i++
			default:
				if matches := excluder.FindStringSubmatch(a); len(matches) == 3 {
					exclude = matches[2]
				} else {
					exit("unknown command line option %q", a)
				}
			}
		default:
			if runtime.GOOS == "windows" {
				if filePattern == "" {
					filePattern = a
				} else {
					files = append(files, a)
				}
			} else {
				files = append(files, a)
			}
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
