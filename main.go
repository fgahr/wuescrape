package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/pkg/errors"
)

type searchResult struct {
	number         string
	title          string
	kind           string
	respTeacher    string
	execTeacher    string
	unit           string
	sws            string
	links          string
	parallelGroups string
	detailLink     string
}

func writeResultHeader(w io.Writer) {
	fmt.Fprintf(w, "%s|%s|%s|%s|%s|%s\n", "Nummer", "Veranstaltungstitel", "Veranstaltungsart", "Dozent/-in (verantw.)", "Dozent/-in (durchf.)", "Organisationseinheit")
}

func (r *searchResult) writeAsCSVRow(w io.Writer) {
	fmt.Fprintf(w, "%s|%s|%s|%s|%s|%s\n", r.number, r.title, r.kind, r.respTeacher, r.execTeacher, r.unit)
}

type semester struct {
	kind int // summer = 1, winter = 2
	year int
}

func parseSemester(s string) (semester, error) {
	var year int
	var kind int
	var semesterIndicator string
	n, err := fmt.Sscanf(s, "%4d%s", &year, &semesterIndicator)
	if err != nil || n < 2 {
		return semester{}, errors.Errorf("malformed argument: %s", s)
	}

	semesterIndicator = strings.ToLower(semesterIndicator)
	if semesterIndicator == "s" {
		kind = 1
	} else if semesterIndicator == "w" {
		kind = 2
	} else {
		return semester{}, errors.Errorf("malformed argument: %s", s)
	}

	return semester{kind, year}, nil
}

func (s semester) fmtSelect() string {
	return fmt.Sprintf("eq|%d|%d", s.kind, s.year)
}

func (s semester) fmtSelectInput() string {
	var sem string
	if s.kind == 1 {
		sem = "Sommersemester"
	} else if s.kind == 2 {
		sem = "Wintersemester"
	} else {
		log.Fatalf("invalid semester: %v", s)
	}
	return fmt.Sprintf("%s+%d", sem, s.year)
}

func usage() {
	fmt.Fprintf(os.Stderr, "%s: [pattern] [semester]\n", os.Args[0])
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "    pattern      any valid wuestudy search pattern")
	fmt.Fprintln(os.Stderr, "    semester     yyyy(s|w), e.g. 2020W for winter semester 2020")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "NOTE: only 300 search results can be shown")
}

func run() error {
	if len(os.Args) < 3 {
		usage()
		os.Exit(1)
	}
	searchTerm := os.Args[1]
	semester := os.Args[2]

	sem, err := parseSemester(semester)
	if err != nil {
		return err
	}

	s := newSession(true)
	s.establish()

	s.submitSearch(searchTerm, sem)
	s.getSearchResultDocument()
	results := s.extractResultData()
	s.addDetailInfo(results)
	if s.err != nil {
		return s.err
	}

	if len(results) == 0 {
		return errors.New("no results found")
	}

	writeResultHeader(os.Stdout)
	for _, r := range results {
		r.writeAsCSVRow(os.Stdout)
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
