package crawler

import (
	"log"
	net "net/url"
	"strings"

	"golang.org/x/net/html"

	"robot/bro"
)

type Client interface {
	RequestUrl(url string, callback bro.DocumentCallbackFunc) (int, error)
}

type Parser interface {
	Parse(status int, url string, doc *html.Node) ([]string, []string)
}

type Listener interface {
	On(status int, url string, foundUrls []string, foundErrors []string)
}

type Crawler struct {
	Site                string
	MaxParallelRequests int
	MaxDepth            int
	Client              Client
	Parser              Parser
	Listener            Listener
	LogLevel            int
}

type Task struct {
	urls  []string
	depth int
}

//
// Normalize trasforms urls to absolute one if necessary.
// Query parameters will be removed except 'uuid' and 'activeComponent'. Remained parameters will be sorted.
// Fragments will be removed. If url is empty, '#', or can not be parsed then method returns nil.
// Note. Relative url is supposed to be with base equals site host even it not begins with slash.
//
// Examples:
//		www.example.com/user, www.example.com/images?name=holidays -> www.example.com/images
//		www.example.com/user, /images?name=holidays -> www.example.com/images
//		www.example.com/user, images?name=holidays -> www.example.com/images
func normalize(site *net.URL, url string) *net.URL {

	if url == "" || strings.HasPrefix(url, "#") {
		return nil
	}

	parsed, err := net.Parse(url)

	if err != nil {
		return nil
	}

	if !parsed.IsAbs() {
		parsed = site.ResolveReference(parsed)
	}

	params := parsed.Query()

	if len(params) > 0 {

		newParams := net.Values{}

		if uuid := params.Get("uuid"); uuid != "" {
			newParams.Set("uuid", uuid)
		}

		if activeComponent := params.Get("activeComponent"); activeComponent != "" {
			newParams.Set("activeComponent", activeComponent)
		}

		parsed.RawQuery = newParams.Encode()
	}

	parsed.Fragment = ""

	return parsed
}

//
// Validate if url is internal and not empty
//
func validate(site, url *net.URL) bool {
	return url != nil &&
		url.Scheme == site.Scheme &&
		url.Host == site.Host &&
		url.Path != "/fx/" &&
		!strings.Contains(url.String(), "fckdsh")
}

func (self *Crawler) visit(url string) ([]string, error) {
	var urls, errors []string
	status, err := self.Client.RequestUrl(url, func(status int, url string, doc *html.Node) {
		urls, errors = self.Parser.Parse(status, url, doc)
	})
	if err != nil {
		return nil, err
	}
	if self.Listener != nil {
		self.Listener.On(status, url, urls, errors)
	}
	return urls, nil
}

func (self *Crawler) request(id int, url string, depth int, queue chan Task, pool chan bool) {

	defer func() {
		<-pool
	}()

	pool <- true

	urls, err := self.visit(url)

	if err != nil {
		log.Printf("[%d] error - %s\n", id, err)
		queue <- Task{nil, depth + 1}
		return
	}

	if self.LogLevel > 0 {
		log.Printf("[%d] %s, depth: %d, more: %d\n", id, url, depth, len(urls))
	}

	queue <- Task{urls, depth + 1}
}

func (self *Crawler) Start(url string) int {

	var visited map[string]bool
	var queue chan Task
	var pool chan bool

	siteUrl, err := net.Parse(self.Site)

	if err != nil {
		return 0
	}

	visited = make(map[string]bool)
	queue = make(chan Task)
	pool = make(chan bool, self.MaxParallelRequests)

	id := 0
	waiting := 0

	newRequest := func(nextUrl string, depth int) {

		resolved := normalize(siteUrl, nextUrl)

		if !validate(siteUrl, resolved) {
			return
		}

		url := resolved.String()

		if !visited[url] {

			id++
			waiting++
			visited[url] = true

			go self.request(id, url, depth, queue, pool)

		}

	}

	newRequest(url, 0)

	for waiting > 0 {

		task := <-queue

		waiting--

		if task.urls != nil && (self.MaxDepth == -1 || task.depth <= self.MaxDepth) {

			for _, url := range task.urls {

				newRequest(url, task.depth)

			}
		}

	}

	return len(visited)

}
