package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"

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
	flowID            int
	authenticityToken string
	cookies           []*http.Cookie
	client            *http.Client
	debug             bool
}

func newSession(debug bool) *session {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatal(errors.Wrap(err, "unable to create session cookie jar"))
	}
	return &session{nil, 1, "", nil, &http.Client{Jar: jar}, debug}
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
		"_flowExecutionKey": {"e1s1"},
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

	submitSearch = "genericSearchMask_SUBMIT"
	viewState    = "javax.faces.ViewState"

	searchMaskID = "genericSearchMask:search_e4ff321960e251186ac57567bec9f4ce:cm_exa_eventprocess_basic_data:fieldset:inputField_0_1ad08e26bde39c9e4f1833e56dcce9b5:id1ad08e26bde39c9e4f1833e56dcce9b5"

	// TODO: necessary?
	navi1NumRows   = "genSearchRes:id3df798d58b4bacd9:id3df798d58b4bacd9NaviNumRowsInput"
	navi2NumRows   = "genSearchRes:id3df798d58b4bacd9:id3df798d58b4bacd9Navi2NumRowsInput"
	naviAboveTable = "genSearchRes:id3df798d58b4bacd9:j_id_5q_l_hk_11:j_id_5q_l_hk_83:j_id_5q_l_hk_87"
	naviBelowTable = "genSearchRes:id3df798d58b4bacd9:j_id_5q_l_hk_11:j_id_5q_l_hk_83:j_id_5q_l_hk_89"
	tableColumns   = "genSearchRes:id3df798d58b4bacd9:j_id_5q_l_hk_11:j_id_5q_l_hk_1b:cols"

	tablePageSize = "genSearchRes:id3df798d58b4bacd9:j_id_5q_l_hk_11:j_id_5q_l_hk_ar:defaultTablePageSize"
)

func (s *session) submitSearch(input string, sem semester) {
	if s.err != nil {
		return
	}

	s.printDebugInfo()

	// POST is forwarded to a GET request with the search results
	resp, err := s.postForm(s.flowURL(), url.Values{
		"activePageElementId":    {searchMaskID},
		"refreshButtonClickedId": {""},
		"navigationPosition":     {"studiesOffered,searchCourses"},
		authToken:                {s.authenticityToken},
		"autoScroll":             {"0,0"},
		navi1NumRows:             {"300"},
		navi2NumRows:             {"300"},
		searchField:              {input},
		termSelectField:          {sem.fmtSelect()},
		termSelctInputField:      {sem.fmtSelectInput()},
		tableColumns: {
			"ActionsBefore",
			"sul.common.Unit.elementnr",
			"sul.plan.searchLecture.veranstTitle",
			"sul.common.Course.eventtypeId",
			"cm.exa.eventprocess.responsible_instructor",
			"cm.exa.eventprocess.instructor",
			"cm.exa.Unit.Orgunit",
			"ActionsAfter",
		},
		naviAboveTable:            {"true"},
		naviBelowTable:            {"false"},
		tablePageSize:             {"300"},
		submitSearch:              {"1"},
		viewState:                 {"e1s1"},
		"SCROLL_TO_ANCHOR":        {},
		"DISABLE_AUTOSCROLL":      {"true"},
		"DISABLE_VALIDATION":      {"true"},
		"genericSearchMask:_idcl": {"genericSearchMask:buttonsBottom:search"},
		"genSearchRes:_idcl":      {"genSearchRes:id3df798d58b4bacd9:j_id_5q_l_hk_11:save"},
	})
	if err != nil {
		s.err = errors.Wrap(err, "failed to submit search parameters")
		return
	}
	defer resp.Body.Close()

	fmt.Fprintln(os.Stderr, resp.Request)

	text, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		s.err = errors.Wrap(err, "failed to read search results")
		return
	}
	fmt.Println(string(text))
}

func (s *session) viewSearchResults() {
	if s.err != nil {
		return
	}

	s.printDebugInfo()
	resp, err := s.get(s.flowURL())
	if err != nil {
		s.err = errors.Wrap(err, "failed to retrieve search results")
		return
	}
	defer resp.Body.Close()

	text, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		s.err = errors.Wrap(err, "failed to read search results")
		return
	}

	fmt.Println(string(text))
}
