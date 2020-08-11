# A Webscraper for Wuestudy Course Data

## Installation

With the go toolchain set up, simply

```text
go get -u github.com/fgahr/wuescrape
```

## Overview

```text
$ wuescrape
wuescrape [opts] <pattern> <semester>

  pattern
    	any valid wuestudy search pattern
  semester
    	yyyy(s|w), e.g. 2020W for winter semester 2020

Available options:
  -details
    	fetch additional course details (may cause slowdown)
```

## Output

Output is pipe (`|`) separated CSV data with a header row:

```text
$ wuescrape "*" 2020W
Nummer|Veranstaltungstitel|Veranstaltungsart|Dozent/-in (verantw.)|Dozent/-in (durchf.)|Organisationseinheit|SWS|Link
... 300 output lines omitted ...
```
