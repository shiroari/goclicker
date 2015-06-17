package parser

import (
	"fmt"
	"regexp"
	"robot/bro"

	"golang.org/x/net/html"
)

type Parser struct{}

func (self *Parser) Parse(status int, url string, doc *html.Node) ([]string, []string) {

	var res, errors []string

	if status != 200 {
		return res, []string{"Broken page"}
	}

	loginForm := bro.First(bro.GetElementsById(doc, "LogonForm"))

	if loginForm != nil {
		return res, []string{fmt.Sprintf("Login form detected at %s", url)}
	}

	errors = findErrors(doc)
	res = findUrls(doc)

	return res, errors
}

func findUrls(doc *html.Node) []string {

	links := bro.GetElementsByTag(doc, "a")
	res := make([]string, 0, links.Len())

	for e := links.Front(); e != nil; e = e.Next() {

		if href := bro.GetAttribute(e.Value.(*html.Node), "href"); href != "" && href != "#" {
			res = append(res, href)
		}

		if onclick := bro.GetAttribute(e.Value.(*html.Node), "onclick"); onclick != "" {
			if jsHref := parseOnClick(onclick); jsHref != "" {
				res = append(res, jsHref)
			}
		}

	}

	return res
}

func findErrors(doc *html.Node) []string {

	trace := bro.First(bro.GetElementsById(doc, "stackTrace"))

	if trace != nil {
		return []string{"Broken page"}
	}

	error1 := bro.First(bro.GetElementsByClass(doc, "error"))

	if error1 != nil {
		return []string{bro.GetText(error1)}
	}

	error2 := bro.First(bro.GetElementsByClass(doc, "message-error"))

	if error2 != nil {
		return []string{bro.GetText(error2)}
	}

	return nil
}

var Confirm = regexp.MustCompile("[^a-zA-Z_]+confirm\\s*\\(")
var Browser = regexp.MustCompile("window\\.open\\s*\\(\\s*['\"]+([^'\"]*)")
var GUIC = regexp.MustCompile("open_in_new_window\\s*\\(\\s*['\"]+([^'\"]*)")

func parseOnClick(str string) string {
	if Confirm.MatchString(str) {
		return ""
	}
	if res := Browser.FindStringSubmatch(str); len(res) > 1 {
		return res[1]
	}
	if res := GUIC.FindStringSubmatch(str); len(res) > 1 {
		return res[1]
	}
	return ""
}
