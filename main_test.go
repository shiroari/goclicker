package main

import (
	"fmt"
	"testing"
)

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
	fmt.Println(url)
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
