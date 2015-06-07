/**
 * Simple Crawler Tool
 *
 * TODO:
 *	  - collect links inside inline javascript (onclick)
 *    - detect long waiting pages
 *    - SIGTERM
 **/

package main

import (
	"log"
	net "net/url"
	"os"
	"robot/bro"
	"runtime"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type Client interface {
	GetUrls(url string) (urls []string, err error)
}

type Task struct {
	urls  []string
	depth int
}

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

func validate(site, url *net.URL) bool {
	return url != nil &&
		url.Scheme == site.Scheme &&
		url.Host == site.Host &&
		url.Path != "/fx/" &&
		!strings.Contains(url.String(), "fckdsh")
}

func visit(url string, client Client) ([]string, error) {
	urls, err := client.GetUrls(url)
	if err != nil {
		return nil, err
	}
	return urls, nil
}

func request(id int, url string, depth int, client Client, queue chan Task, pool chan bool) {

	defer func() {
		<-pool
	}()

	pool <- true

	urls, err := visit(url, client)

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

func crawl(site, url string, maxParallelRequests, maxDepth int, client Client) int {

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

			go request(id, url, depth, client, queue, pool)

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

func findErrorPage(url string, status int, doc *html.Node) {

	if status != 200 {
		logPageError(url, status, "Broken page")
		return
	}

	loginForm := bro.First(bro.GetElementsById(doc, "LogonForm"))

	if loginForm != nil {
		logPageError(url, status, "Login form detected")
		os.Exit(1)
	}

	trace := bro.First(bro.GetElementsById(doc, "stackTrace"))

	if trace != nil {
		logPageError(url, status, "Broken page")
		return
	}

	error1 := bro.First(bro.GetElementsByClass(doc, "error"))

	if error1 != nil {
		logPageError(url, status, bro.GetText(error1))
	}

	error2 := bro.First(bro.GetElementsByClass(doc, "message-error"))

	if error2 != nil {
		logPageError(url, status, bro.GetText(error2))
	}
}

func logPageError(url string, status int, message string) {
	log.Printf("Error page %d :: %s :: %s\n", status, url, message)
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

	runtime.GOMAXPROCS(2)

	defer showStatistics(readStats())

	logLevel = 1
	maxParallelRequests := 20
	maxDepth := -1

	client := bro.New(0, findErrorPage)

	visited := crawl("http://localhost:8080", "/fx/auth", maxParallelRequests, maxDepth, client)

	log.Printf("Visited %d url(s)", visited)
}
