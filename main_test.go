package main

import (
	"fmt"
	_url_ "net/url"
	"testing"
)

func TestNormalizeShouldSkipInvalidURLs(t *testing.T) {

	root, _ := _url_.Parse("http://localhost:8080/")

	if res := normalize(root, ""); res != "" {
		t.Errorf("normalize should skip empty url")
	}

	if res := normalize(root, "/"); res != "http://localhost:8080/" {
		t.Errorf("normalize should handle root url")
	}

	if res := normalize(root, "#"); res != "" {
		t.Errorf("normalize should skip url with empty fragment")
	}

	if res := normalize(root, "#fragment"); res != "" {
		t.Errorf("normalize should skip url with any fragment")
	}

	if res := normalize(root, "javascript:void(0)"); res != "" {
		t.Errorf("normalize should skip unexpected scheme")
	}
}

func TestNormalizeShouldKeepOriginalHost(t *testing.T) {

	root, _ := _url_.Parse("http://localhost:8080/")

	if res := normalize(root, "http://host:8080/fx/"); res != "http://host:8080/fx/" {
		t.Errorf("normalize should keep original host")
	}
}

func TestNormalizeShouldFixPath(t *testing.T) {

	root, _ := _url_.Parse("http://localhost:8080/")

	if res := normalize(root, "/fx/"); res != "http://localhost:8080/fx/" {
		t.Errorf("normalize should fix path: %s", res)
	}

	if res := normalize(root, "fx/"); res != "http://localhost:8080/fx/" {
		t.Errorf("normalize should fix path: %s", res)
	}

	if res := normalize(root, "fx"); res != "http://localhost:8080/fx" {
		t.Errorf("normalize should fix path: %s", res)
	}

	root2, _ := _url_.Parse("http://localhost:8080")

	if res := normalize(root2, "fx"); res != "http://localhost:8080/fx" {
		t.Errorf("normalize should fix path: %s", res)
	}

}

func TestNormalizeShouldDropFragment(t *testing.T) {

	root, _ := _url_.Parse("http://localhost:8080/")

	if res := normalize(root, "jsp#id"); res != "http://localhost:8080/jsp" {
		t.Errorf("normalize should drop any fragment")
	}

}

func TestNormalizeShouldDropNotRelevantQueryParamsAndSortOthers(t *testing.T) {

	root, _ := _url_.Parse("http://localhost:8080/")

	should1 := "http://localhost:8080/?activeComponent=Reports&uuid=coreboo2k02bo0000kfd4plquvalkl6k"
	if res := normalize(root, "?uuid=coreboo2k02bo0000kfd4plquvalkl6k&activeComponent=Reports"); res != should1 {
		t.Errorf("normalize should sort query params: %s", res)
	}

	if res := normalize(root, "http://localhost:8080/?uuid=coreboo2k02bo0000kfd4plquvalkl6k&activeComponent=Reports"); res != should1 {
		t.Errorf("normalize should sort query params: %s", res)
	}

	should2 := "http://localhost:8080/?activeComponent=Reports"
	if res := normalize(root, "?activeComponent=Reports&dropme=yes"); res != should2 {
		t.Errorf("normalize should drop not relevant query params: %s", res)
	}

	if res := normalize(root, "http://localhost:8080/?activeComponent=Reports&dropme=yes"); res != should2 {
		t.Errorf("normalize should drop not relevant query params: %s", res)
	}
}

func TestCrawlWithoutDeepLimit(t *testing.T) {

	should := 5

	visited := crawl("http://golang.org", "/", 2, -1, fetcher)

	if visited != should {
		t.Errorf("Test shoud visit %d urls but was %d", should, visited)
	}
}

func TestCrawlWithZeroDeepLimit(t *testing.T) {

	should := 1

	visited := crawl("http://golang.org", "/", 2, 0, fetcher)

	if visited != should {
		t.Errorf("Test shoud visit %d urls but was %d", should, visited)
	}
}

func TestCrawlWithDeepLimit(t *testing.T) {

	should := 3

	visited := crawl("http://golang.org", "/", 2, 1, fetcher)

	if visited != should {
		t.Errorf("Test shoud visit %d urls but was %d", should, visited)
	}
}

type webpage map[string][]string

func (self webpage) GetUrls(url string) ([]string, error) {
	if res, ok := self[url]; ok {
		return res, nil
	}
	return nil, fmt.Errorf("not found: %s", url)
}

var fetcher = webpage{
	"http://golang.org/": []string{
		"/pkg/",
		"/cmd/",
	},
	"http://golang.org/pkg/": []string{
		"/",
		"/cmd/",
		"/pkg/fmt/",
		"/pkg/os/",
	},
	"http://golang.org/pkg/fmt/": []string{
		"/",
		"/pkg/",
	},
	"http://golang.org/pkg/os/": []string{
		"/",
		"/pkg/",
	},
}
