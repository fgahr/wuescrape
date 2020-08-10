package main

import (
	"fmt"
	"os"
)

type semester struct {
	winter bool
	year   int
}

func parseSemester(s string) (semester, error) {
	// TODO
	return semester{true, 2020}, nil
}

func (s semester) fmtSelect() string {
	return "eq|2|2020"
}

func (s semester) fmtSelectInput() string {
	// TODO
	return "Wintersemester+2020"
}

func run() error {
	searchTerm := os.Args[1]
	semester := os.Args[2]

	s := newSession(false)
	s.establish()

	sem, err := parseSemester(semester)
	if err != nil {
		return err
	}

	s.submitSearch(searchTerm, sem)
	s.enlargeResultTable()

	return s.err
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
