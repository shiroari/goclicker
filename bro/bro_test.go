package bro

import (
	"testing"
	"strings"
	"golang.org/x/net/html"
)

func Test_getElementsByTag(t *testing.T) {

	fixture := strings.NewReader("<html><body><div/><DIV/></body></html>")

	doc, _ := html.Parse(fixture)
	
	res := getElementsByTag(doc, "div")

	if res.Len() != 2 {
		t.Errorf("List shoud has 2 div elements but was %d", res.Len())	
	}
	
}