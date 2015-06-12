/**
 * Simple Crawler Tool
 *
 * TODO:
 *	  - CLI
 *    - SIGTERM
 *    - Detect long loading pages
 **/

package main

import (
	"log"
	net "net/url"
	"runtime"
	"strings"
	"time"

	"golang.org/x/net/html"

	"robot/bro"
	p "robot/parser"
)

type Client interface {
	RequestUrl(url string, callback bro.DocumentCallbackFunc) (int, error)
}

type Parser interface {
	Parse(status int, url string, doc *html.Node) ([]string, []string)
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

func visit(url string, client Client, parser Parser) ([]string, error) {
	var urls, errors []string
	status, err := client.RequestUrl(url, func(status int, url string, doc *html.Node) {
		urls, errors = parser.Parse(status, url, doc)
	})
	if err != nil {
		return nil, err
	}
	if errors != nil {
		logPageError(url, status, errors)
	}
	return urls, nil
}

func request(id int, url string, depth int, client Client, parser Parser, queue chan Task, pool chan bool) {

	defer func() {
		<-pool
	}()

	pool <- true

	urls, err := visit(url, client, parser)

	if err != nil {
		log.Printf("[%d] error - %s\n", id, err)
		queue <- Task{nil, depth + 1}
		return
	}

	if logLevel > 0 {
		log.Printf("[%d] %s, depth: %d, more: %d\n", id, url, depth, len(urls))
	}

	queue <- Task{urls, depth + 1}
}

func crawl(site, url string, maxParallelRequests, maxDepth int, client Client, parser Parser) int {

	var visited map[string]bool
	var queue chan Task
	var pool chan bool

	siteUrl, err := net.Parse(site)

	if err != nil {
		return 0
	}

	visited = make(map[string]bool)
	queue = make(chan Task)
	pool = make(chan bool, maxParallelRequests)

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

			go request(id, url, depth, client, parser, queue, pool)

		}

	}

	newRequest(url, 0)

	for waiting > 0 {

		task := <-queue

		waiting--

		if task.urls != nil && (maxDepth == -1 || task.depth <= maxDepth) {

			for _, url := range task.urls {

				newRequest(url, task.depth)

			}
		}

	}

	return len(visited)

}

func logPageError(url string, status int, errors []string) {
	for err := range errors {
		log.Printf("Error page %d :: %s :: %s\n", status, url, err)
	}
}

func showStatistics(start time.Time, mem1 *runtime.MemStats) {
	_, mem2 := readStats()
	log.Printf("Number of Goroutines: %d\n", runtime.NumGoroutine())
	log.Printf("Memory stat: %.2f Mb\n", float64(mem2.Sys-mem1.Sys)/(1024*1024))
	log.Printf("Time: %s\n", time.Since(start))
}

func readStats() (time.Time, *runtime.MemStats) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	return time.Now(), &mem
}

var logLevel int

func main() {

	//runtime.GOMAXPROCS(2)

	defer showStatistics(readStats())

	logLevel = 1
	maxParallelRequests := 20
	maxDepth := -1

	client := bro.New("system", "manager", 0)
	parser := p.New()

	visited := crawl("http://localhost:8080", "/fx/auth", maxParallelRequests, maxDepth, client, parser)

	log.Printf("Visited %d url(s)", visited)
}
