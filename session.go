package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

// URLs
const (
	baseURL   = "https://wuestudy.zv.uni-wuerzburg.de"
	searchURL = baseURL + "/qisserver/pages/cm/exa/coursemanagement/basicCourseData.xhtml"
)

func mustParse(s string) *url.URL {
	u, err := url.Parse(searchURL)
	if err != nil {
		log.Fatal(err)
	}
	return u
}

func bURL() *url.URL {
	return mustParse(baseURL)
}

func sURL() *url.URL {
	return mustParse(searchURL)
}

type session struct {
	err               error
	document          *goquery.Document
	flowID            int
	authenticityToken string
	cookies           []*http.Cookie
	client            *http.Client
	debug             bool
}

func newSession(printDebugInfo bool) *session {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatal(errors.Wrap(err, "unable to create session cookie jar"))
	}
	return &session{
		flowID: 1,
		client: &http.Client{Jar: jar},
		debug:  printDebugInfo,
	}
}

func (s *session) printDebugInfo() {
	if !s.debug {
		return
	}

	fmt.Fprintln(os.Stderr, "### DEBUG ##################################")
	fmt.Fprintf(os.Stderr, "error: %v\nauthenticity_token: %s\n",
		s.err, s.authenticityToken)
	fmt.Fprintln(os.Stderr, "## Cookies ##")
	for _, c := range s.client.Jar.Cookies(sURL()) {
		fmt.Fprintln(os.Stderr, c)
	}
	fmt.Fprintln(os.Stderr, "### END DEBUG ##############################")
}

func (s *session) flowExecKey() string {
	return fmt.Sprintf("e1s%d", s.flowID)
}

func (s *session) get(url string) (*http.Response, error) {
	if s.debug {
		fmt.Fprintln(os.Stderr, "### GET ####################################")
		fmt.Fprintln(os.Stderr, url)
		fmt.Fprintln(os.Stderr, "### END GET ################################")
	}
	return s.client.Get(url)
}

func (s *session) postForm(url string, data url.Values) (*http.Response, error) {
	if s.debug {
		fmt.Fprintln(os.Stderr, "### POST ###################################")
		fmt.Fprintln(os.Stderr, url)
		for k, v := range data {
			fmt.Fprintf(os.Stderr, "%s = %s\n", k, v)
		}
		fmt.Fprintln(os.Stderr, "### END POST ###############################")
	}
	return s.client.PostForm(url, data)
}

func (s *session) flowURL() string {
	return fmt.Sprintf("%s?%s", searchURL, url.Values{
		"_flowId":           {"searchCourseNonStaff-flow"},
		"_flowExecutionKey": {s.flowExecKey()},
	}.Encode())
}

func (s *session) establish() {
	if s.err != nil {
		return
	}

	resp, err := s.client.Get(s.flowURL())
	if err != nil {
		s.err = errors.Wrap(err, "failed to fetch search page")
		return
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		s.err = errors.Wrap(err, "unable to read response document")
		return
	}

	authTok := doc.Find("input").FilterFunction(func(_ int, s *goquery.Selection) bool {
		name, ok := s.Attr("name")
		return ok && name == authToken
	}).First()
	if authTok == nil {
		s.err = errors.Errorf("no %s field found", authToken)
		return
	}

	authTkn, ok := authTok.Attr("value")
	if !ok {
		s.err = errors.Errorf("%s has no value", authToken)
		return
	}
	s.authenticityToken = authTkn

	s.client.Jar.SetCookies(bURL(), []*http.Cookie{
		&http.Cookie{Name: "download-complete", Value: ""},
		&http.Cookie{Name: "sessionRefresh", Value: "0"},
	})
}

// name/id constants
const (
	authToken = "authenticity_token"

	searchField         = "genericSearchMask:search_e4ff321960e251186ac57567bec9f4ce:cm_exa_eventprocess_basic_data:fieldset:inputField_0_1ad08e26bde39c9e4f1833e56dcce9b5:id1ad08e26bde39c9e4f1833e56dcce9b5"
	termSelectField     = "genericSearchMask:search_e4ff321960e251186ac57567bec9f4ce:cm_exa_eventprocess_basic_data:fieldset:inputField_3_abb156a1126282e4cf40d48283b4e76d:idabb156a1126282e4cf40d48283b4e76d:termSelect"
	termSelctInputField = "genericSearchMask:search_e4ff321960e251186ac57567bec9f4ce:cm_exa_eventprocess_basic_data:fieldset:inputField_3_abb156a1126282e4cf40d48283b4e76d:idabb156a1126282e4cf40d48283b4e76d:termSelectInput"

	submitSearchMask = "genericSearchMask_SUBMIT"
	submitSearchRes  = "genSearchRes_SUBMIT"
	viewState        = "javax.faces.ViewState"

	navi1NumRows = "genSearchRes:id3df798d58b4bacd9:id3df798d58b4bacd9NaviNumRowsInput"
	navi2NumRows = "genSearchRes:id3df798d58b4bacd9:id3df798d58b4bacd9Navi2NumRowsInput"

	tablePageSize = "genSearchRes:id3df798d58b4bacd9:j_id_5q_l_hk_11:j_id_5q_l_hk_ar:defaultTablePageSize"
)

func (s *session) submitSearch(input string, sem semester) {
	if s.err != nil {
		return
	}

	s.printDebugInfo()

	resp, err := s.postForm(s.flowURL(), url.Values{
		authToken:                 {s.authenticityToken},
		searchField:               {input},
		termSelectField:           {sem.fmtSelect()},
		termSelctInputField:       {sem.fmtSelectInput()},
		submitSearchMask:          {"1"},
		viewState:                 {s.flowExecKey()},
		"genericSearchMask:_idcl": {"genericSearchMask:buttonsBottom:search"},
	})
	if err != nil {
		s.err = errors.Wrap(err, "failed to submit search parameters")
		return
	}
	defer resp.Body.Close()

	s.flowID++
}

func (s *session) getSearchResultDocument() {
	if s.err != nil {
		return
	}

	s.printDebugInfo()

	resp, err := s.postForm(s.flowURL(), url.Values{
		authToken: {s.authenticityToken},
		// 300 is the maximum result table size
		navi1NumRows:    {"300"},
		navi2NumRows:    {"300"},
		tablePageSize:   {"300"},
		submitSearchRes: {"1"},
		viewState:       {s.flowExecKey()},
	})
	if err != nil {
		s.err = errors.Wrap(err, "failed to submit table preferences")
		return
	}
	defer resp.Body.Close()

	s.document, s.err = goquery.NewDocumentFromResponse(resp)
}

func plainContent(td *goquery.Selection) string {
	// td.RemoveFiltered("span")
	return td.Text()
}

func (s *session) extractResultData() []*searchResult {
	if s.err != nil {
		return nil
	}

	if s.document == nil {
		log.Fatal("document missing but no error encountered")
	}

	tableBody := s.document.Find("table").FilterFunction(func(_ int, sel *goquery.Selection) bool {
		return sel.AttrOr("id", "") == "genSearchRes:id3df798d58b4bacd9:id3df798d58b4bacd9Table"
	}).Find("tbody")

	resultData := make([]*searchResult, 0)
	tableBody.Find("tr").Each(func(_ int, tds *goquery.Selection) {
		result := searchResult{}
		tds.Find("td").Each(func(_ int, td *goquery.Selection) {
			class, ok := td.Attr("class")
			if !ok {
				return
			}

			switch class {
			case "smallestPossible singleLine column0":
				result.detailLink = td.Find("a").AttrOr("href", "")
			case " column1":
				result.number = plainContent(td)
			case " column2":
				result.title = td.Find("button").Find("span").Text()
			case " column3":
				result.kind = plainContent(td)
			case " column4":
				result.respTeacher = plainContent(td)
			case " column5":
				result.execTeacher = plainContent(td)
			case " column6":
				result.unit = plainContent(td)
			}
		})
		resultData = append(resultData, &result)
	})

	return resultData
}

func (s *session) addDetailInfo(results []*searchResult) {
	concurrencyLimit := 50
	semaphore := make(chan struct{}, concurrencyLimit)
	done := make(chan struct{})
	started := 0

	for _, res := range results {
		go func(r *searchResult) {
			semaphore <- struct{}{}
			s.addResultDetail(r)
			<-semaphore
			done <- struct{}{}
		}(res)
		started++
	}

	// Make sure all goroutines are finished
	for i := 0; i < started; i++ {
		<-done
	}
}

func (s *session) addResultDetail(res *searchResult) {
	if res.detailLink == "" {
		return
	}

	// detail links are of the form /qisserver/..
	resp, err := s.get(baseURL + res.detailLink)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		log.Println(err)
		return
	}

	res.sws = doc.Find("div").FilterFunction(func(_ int, sel *goquery.Selection) bool {
		return sel.AttrOr("id", "") == "109dced638ab3377d8214df3f0097fdd"
	}).First().Text()
	// This field has been known to include gratuitous linebreaks
	res.sws = strings.TrimSpace(res.sws)

	res.link = doc.Find("a").FilterFunction(func(_ int, sel *goquery.Selection) bool {
		href, ok := sel.Attr("href")
		if !ok {
			return false
		}

		// Non-course links are relative and start with a slash
		return strings.HasPrefix(href, "https://") &&
			strings.Contains(href, "uni-wuerzburg.de") &&
			// For course links the full link is included as text as well
			strings.TrimSpace(href) == strings.TrimSpace(sel.Text())
	}).First().AttrOr("href", "")
	res.link = strings.TrimSpace(res.link)
}
