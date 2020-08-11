# A Webscraper for Wuestudy Course Data

## Overview

```text
$ wuescrape
wuescrape: [pattern] [semester]

    pattern      any valid wuestudy search pattern
    semester     yyyy(s|w), e.g. 2020W for winter semester 2020

NOTE: only 300 search results can be shown
```

The 300 result limitation is -- sadly -- real for now due to the heavy use
of javascript in the way wuestudy handles search data.

## Output

Output is pipe (`|`) separated CSV data with a header row:

```text
$ wuescrape "*" 2020W
Nummer|Veranstaltungstitel|Veranstaltungsart|Dozent/-in (verantw.)|Dozent/-in (durchf.)|Organisationseinheit
... 300 output lines omitted ...
```
