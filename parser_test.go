package main

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestParserShouldSkipErrorPages(t *testing.T) {

	fixture := strings.NewReader("<html><body><a href='url:test1'><a href='url:test2'></body></html>")

	doc, _ := html.Parse(fixture)

	parser := Parser{}

	res, errors := parser.Parse(500, "test", doc)

	if len(errors) != 1 {
		t.Error("Error not found")
	}

	if res != nil {
		t.Error("Links should be ignored")
	}

}

func TestParserShouldFindErrorsOnPage(t *testing.T) {

	fixture := strings.NewReader("<html><body><div id='stackTrace'></div></body></html>")

	doc, _ := html.Parse(fixture)

	parser := Parser{}

	_, errors := parser.Parse(200, "test", doc)

	if len(errors) != 1 {
		t.Error("Error not found")
	}

}

func TestParserShouldFindLoginPage(t *testing.T) {

	fixture := strings.NewReader("<html><body><form id='LogonForm'></form></body></html>")

	doc, _ := html.Parse(fixture)

	parser := Parser{}

	_, errors := parser.Parse(200, "test", doc)

	if len(errors) != 1 || !strings.HasPrefix(errors[0], "Login form detected") {
		t.Error("Login form not found")
	}

}

func TestAHrefShouldBeReturned(t *testing.T) {

	fixture := strings.NewReader("<html><body><a href='url:test1'></a><a href='url:test2'></a></body></html>")

	doc, _ := html.Parse(fixture)

	parser := Parser{}

	res, errors := parser.Parse(200, "test", doc)

	if len(res) != 2 {
		t.Errorf("List shoud has 2 urls elements but was %d", len(res))
	}

	if errors != nil {
		t.Error("No errors on page")
	}

	if res[0] != "url:test1" {
		t.Errorf("Parsed url is not correct %s != %s", res[0], "url:test1")
	}

}

func TestOnClickLinkShouldBeReturned(t *testing.T) {

	fixture := strings.NewReader("<html><body>" +
		"<a href='#' onclick=\"window.open('http://host/res/?param=1')\"></a>" +
		"<a href='#' onclick=\"open_in_new_window('http://host/res/?param=1')\"></a>" +
		"<a href='#' onclick=\"open_in_new_window (  '/res?param=1' )\"></a>" +
		"<a href='#' onclick='open_in_new_window(\"/res/?param=1\")'></a>" +
		"<a href='#' onclick=\"window.confirm('sure?'); open_in_new_window('/res/?param=1')\"></a>" +
		"</body></html>")

	doc, _ := html.Parse(fixture)

	parser := Parser{}

	res, errors := parser.Parse(200, "test", doc)

	if errors != nil {
		t.Error("No errors on page")
	}

	if len(res) != 4 {
		t.Errorf("List shoud has 4 urls elements but was %d", len(res))
	}

	if res[0] != "http://host/res/?param=1" {
		t.Errorf("url[0]: %s", res[0])
	}

	if res[1] != "http://host/res/?param=1" {
		t.Errorf("url[1]: %s", res[1])
	}

	if res[2] != "/res?param=1" {
		t.Errorf("url[2]: %s", res[2])
	}

	if res[3] != "/res/?param=1" {
		t.Errorf("url[3]: %s", res[3])
	}
}
