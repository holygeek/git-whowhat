package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/holygeek/piper"
	"log"
	"os"
	"sort"
	"strings"
)

const WHO = "  WHO:"

func main() {
	type Author string
	type Filename string
	type AuthorSet map[Author]bool

	flag.Usage = showHelp

	help := flag.Bool("h", false, "Show usage")
	debug := flag.Bool("d", false, "Debug")
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	var currAuthor Author
	whatwho := make(map[Filename]AuthorSet)

	lineScanner := runGitLog(debug)
	for lineScanner.Scan() {
		line := lineScanner.Text()
		if line == "" {
			continue
		}

		if isAuthorLine(&line) {
			currAuthor = Author(line[len(WHO):])
			continue
		}

		fname := Filename(line)
		if authorSet, ok := whatwho[fname]; ok {
			authorSet[currAuthor] = true
		} else {
			whatwho[fname] = AuthorSet{currAuthor: true}
		}
	}

	authorsAndFiles := make(map[string][]string)
	for fname, authorSet := range whatwho {
		authors := make([]string, 0)
		for author, _ := range authorSet {
			authors = append(authors, string(author))
		}
		sort.Strings(authors)
		key := strings.Join(authors, ",")
		if filelist, ok := authorsAndFiles[key]; ok {
			authorsAndFiles[key] = append(filelist, string(fname))
		} else {
			authorsAndFiles[key] = []string{string(fname)}
		}
	}

	l := len(authorsAndFiles)
	c := 0
	for authors, files := range authorsAndFiles {
		c++
		fmt.Print(strings.Replace(authors, ",", "\n", -1))
		fmt.Print("\n\t")
		sort.Strings(files)
		fmt.Println(strings.Join(files, "\n\t"))
		if c < l {
			fmt.Println()
		}
	}
}

func runGitLog(debug *bool) *bufio.Scanner {
	gitargs := setupGitLogArgs()
	if *debug {
		fmt.Println("git", strings.Join(gitargs, " "))
	}

	stdout, stderr := piper.MustPipe("git", gitargs...)

	go func() {
		for stderr.Scan() {
			fmt.Fprintln(os.Stderr, stderr.Text())
		}
	}()

	return stdout
}

func setupGitLogArgs() []string {
	gitargs := make([]string, 0)
	gitargs = append(gitargs, "log", "--format="+WHO+"%an", "--name-only")
	if flag.NArg() > 0 {
		gitargs = append(gitargs, flag.Args()...)
	} else {
		gitargs = append(gitargs, "ORIG_HEAD..")
	}
	return gitargs
}

func isAuthorLine(line *string) bool {
	return len(*line) > len(WHO) && (*line)[0:len(WHO)] == WHO
}

func showHelp() {
	usage := `NAME
	git-whowhat - Show authors and the files that they modified.

SYNOPSIS
	git whowhat [<options>] [<since>..<until> [[--] <path>...]

OPTIONS
	-d
	    Print debugging information

	-h
	    Show this help message

	[<since>..<until> [[--] <path>...]
	    These are the same argument understood by git log

	    If none is specified, "ORIG_HEAD.." is used as the sole argument
`
	fmt.Print(usage)
}

func die(reason ...interface{}) {
	message := fmt.Sprint("git-whowhat", reason)
	log.Fatal(message)
}
