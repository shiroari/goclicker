package crawler

import (
	"fmt"
	net "net/url"
	"testing"

	"golang.org/x/net/html"

	"robot/bro"
)

func TestValidateShouldSkipInvalidURLs(t *testing.T) {

	root, _ := net.Parse("http://localhost:8080/")

	if validate(root, nil) {
		t.Errorf("validate should skip nil")
	}

	if url, _ := net.Parse("host:8080"); validate(root, url) {
		t.Errorf("validate should skip different domain")
	}

	if url, _ := net.Parse("javascript:void(0)"); validate(root, url) {
		t.Errorf("validate should skip unexpected scheme")
	}

}

func TestNormalizeShouldSkipInvalidURLs(t *testing.T) {

	root, _ := net.Parse("http://localhost:8080/")

	if res := normalize(root, ""); res != nil {
		t.Errorf("normalize should skip empty url")
	}

	if res := normalize(root, "/"); res.String() != "http://localhost:8080/" {
		t.Errorf("normalize should handle root url")
	}

	if res := normalize(root, "#"); res != nil {
		t.Errorf("normalize should skip url with empty fragment")
	}

	if res := normalize(root, "#fragment"); res != nil {
		t.Errorf("normalize should skip url with any fragment")
	}
}

func TestNormalizeShouldKeepOriginalHost(t *testing.T) {

	root, _ := net.Parse("http://localhost:8080/")

	if res := normalize(root, "http://host:8080/fx/"); res.String() != "http://host:8080/fx/" {
		t.Errorf("normalize should keep original host")
	}
}

func TestNormalizeShouldFixPath(t *testing.T) {

	root, _ := net.Parse("http://localhost:8080/")

	if res := normalize(root, "/fx/"); res.String() != "http://localhost:8080/fx/" {
		t.Errorf("normalize should fix path: %s", res)
	}

	if res := normalize(root, "fx/"); res.String() != "http://localhost:8080/fx/" {
		t.Errorf("normalize should fix path: %s", res)
	}

	if res := normalize(root, "fx"); res.String() != "http://localhost:8080/fx" {
		t.Errorf("normalize should fix path: %s", res)
	}

	root2, _ := net.Parse("http://localhost:8080")

	if res := normalize(root2, "fx"); res.String() != "http://localhost:8080/fx" {
		t.Errorf("normalize should fix path: %s", res)
	}

}

func TestNormalizeShouldDropFragment(t *testing.T) {

	root, _ := net.Parse("http://localhost:8080/")

	if res := normalize(root, "jsp#id"); res.String() != "http://localhost:8080/jsp" {
		t.Errorf("normalize should drop any fragment")
	}

}

func TestNormalizeShouldDropNotRelevantQueryParamsAndSortOthers(t *testing.T) {

	root, _ := net.Parse("http://localhost:8080/")

	should1 := "http://localhost:8080/?activeComponent=Reports&uuid=coreboo2k02bo0000kfd4plquvalkl6k"
	if res := normalize(root, "?uuid=coreboo2k02bo0000kfd4plquvalkl6k&activeComponent=Reports"); res.String() != should1 {
		t.Errorf("normalize should sort query params: %s", res)
	}

	if res := normalize(root, "http://localhost:8080/?uuid=coreboo2k02bo0000kfd4plquvalkl6k&activeComponent=Reports"); res.String() != should1 {
		t.Errorf("normalize should sort query params: %s", res)
	}

	should2 := "http://localhost:8080/?activeComponent=Reports"
	if res := normalize(root, "?activeComponent=Reports&dropme=yes"); res.String() != should2 {
		t.Errorf("normalize should drop not relevant query params: %s", res)
	}

	if res := normalize(root, "http://localhost:8080/?activeComponent=Reports&dropme=yes"); res.String() != should2 {
		t.Errorf("normalize should drop not relevant query params: %s", res)
	}
}

func TestCrawlWithoutDeepLimit(t *testing.T) {

	should := 5

	visited := crawl("/", mockConfig(-1))

	if visited != should {
		t.Errorf("Test shoud visit %d urls but was %d", should, visited)
	}
}

func TestCrawlWithZeroDeepLimit(t *testing.T) {

	should := 1

	visited := crawl("/", mockConfig(0))

	if visited != should {
		t.Errorf("Test shoud visit %d urls but was %d", should, visited)
	}
}

func TestCrawlWithDeepLimit(t *testing.T) {

	should := 3

	visited := crawl("/", mockConfig(1))

	if visited != should {
		t.Errorf("Test shoud visit %d urls but was %d", should, visited)
	}
}

func mock(maxDepth int) *Config {
	return &Crawler{"http://golang.org", 2, maxDepth, &clientMock{}, &parserMock{}, nil}
}

type parserMock struct{}
type clientMock struct{}

func (self *parserMock) Parse(status int, url string, doc *html.Node) ([]string, []string) {
	return data[url], nil
}

func (self *clientMock) RequestUrl(url string, callback bro.DocumentCallbackFunc) (int, error) {
	if _, ok := data[url]; ok {
		callback(200, url, nil)
		return 200, nil
	}
	return -1, fmt.Errorf("not found: %s", url)
}

var data = map[string][]string{
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
