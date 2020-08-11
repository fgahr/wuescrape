# A Webscraper for Wuestudy Course Data

## Installation

With the go toolchain set up, simply

```text
go get -u github.com/fgahr/wuescrape
```

## Overview

```text
$ wuescrape
wuescrape: [pattern] [semester]

    pattern      any valid wuestudy search pattern
    semester     yyyy(s|w), e.g. 2020W for winter semester 2020

NOTE: only 300 search results can be shown
```

The 300 result limitation is currently real but may be resolved soon.

## Output

Output is pipe (`|`) separated CSV data with a header row:

```text
$ wuescrape "*" 2020W
Nummer|Veranstaltungstitel|Veranstaltungsart|Dozent/-in (verantw.)|Dozent/-in (durchf.)|Organisationseinheit
... 300 output lines omitted ...
```
