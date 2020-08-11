package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/pkg/errors"
)

type searchResult struct {
	// Whether details have been added, not for output
	detailsAdded   bool
	number         string
	title          string
	kind           string
	respTeacher    string
	execTeacher    string
	unit           string
	sws            string
	link           string
	parallelGroups string
	// The link to the details page, not for output
	detailPageLink string
}

func writeResultHeader(w io.Writer, detailsAdded bool) {
	fmt.Fprintf(w, "%s|%s|%s|%s|%s|%s",
		"Nummer", "Veranstaltungstitel", "Veranstaltungsart",
		"Dozent/-in (verantw.)", "Dozent/-in (durchf.)", "Organisationseinheit")
	if detailsAdded {
		fmt.Fprintf(w, "|%s|%s\n", "SWS", "Link")
	} else {
		fmt.Fprintln(w)
	}
}

func (r *searchResult) writeAsCSVRow(w io.Writer) {
	fmt.Fprintf(w, "%s|%s|%s|%s|%s|%s",
		r.number, r.title, r.kind, r.respTeacher, r.execTeacher, r.unit)
	if r.detailsAdded {
		fmt.Fprintf(w, "|%s|%s\n", r.sws, r.link)
	} else {
		fmt.Fprintln(w)
	}
}

const (
	summer = 1
	winter = 2
)

type semester struct {
	kind int
	year int
}

func parseSemester(s string) (semester, error) {
	var year int
	var kind int
	var semesterIndicator string
	n, err := fmt.Sscanf(s, "%4d%s", &year, &semesterIndicator)
	if err != nil || n < 2 {
		return semester{}, errors.Errorf("malformed semester argument: %s", s)
	}

	semesterIndicator = strings.ToLower(semesterIndicator)
	if semesterIndicator == "s" {
		kind = summer
	} else if semesterIndicator == "w" {
		kind = winter
	} else {
		return semester{}, errors.Errorf("malformed semester argument: %s", s)
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

func printUsage() {
	fmt.Fprintf(os.Stderr, "%s [opts] <pattern> <semester>\n", os.Args[0])
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "  pattern")
	fmt.Fprintln(os.Stderr, "    \tany valid wuestudy search pattern")
	fmt.Fprintln(os.Stderr, "  semester")
	fmt.Fprintln(os.Stderr, "    \tyyyy(s|w), e.g. 2020W for winter semester 2020")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Available options:")
	flag.PrintDefaults()
}

func run() error {
	flag.Usage = printUsage
	detailp := flag.Bool("details", false, "fetch additional course details (may cause slowdown)")
	flag.Parse()
	args := flag.Args()

	if len(args) != 2 {
		printUsage()
		os.Exit(1)
	}
	searchTerm := args[0]
	semester := args[1]

	sem, err := parseSemester(semester)
	if err != nil {
		return err
	}

	s := newSession(false)
	s.establish()

	s.submitSearch(searchTerm, sem)
	s.getSearchResultDocument()
	results := s.extractResultData()
	if *detailp {
		s.addDetails(results)
	}

	if s.err != nil {
		return s.err
	}

	if len(results) == 0 {
		fmt.Fprintln(os.Stderr, "no results found")
		return nil
	}

	writeResultHeader(os.Stdout, *detailp)
	for _, r := range results {
		r.writeAsCSVRow(os.Stdout)
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		printUsage()
		os.Exit(1)
	}
}
