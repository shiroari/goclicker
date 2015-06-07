/**
 * Simple Crawler Tool
 *
 * TODO:
 * 	  - find error pages
 *	  - collect links inside inline javascript (onclick)
 *    - detect long waiting pages
 **/

package main

import (
	"log"
	_url_ "net/url"
	"robot/bro"
	"runtime"
	"strings"
	"time"
)

type Client interface {
	GetUrls(url string) (urls []string, err error)
}

type Task struct {
	urls  []string
	depth int
}

func normalize(root *_url_.URL, url string) string {

	if url == "" || strings.HasPrefix(url, "#") {
		return ""
	}

	parsed, err := _url_.Parse(url)

	if err != nil {
		return ""
	}

	if parsed.IsAbs() {
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			return ""
		}
	} else {
		parsed = root.ResolveReference(parsed)
	}

	params := parsed.Query()

	if len(params) > 0 {

		newParams := _url_.Values{}

		if uuid := params.Get("uuid"); uuid != "" {
			newParams.Set("uuid", uuid)
		}

		if activeComponent := params.Get("activeComponent"); activeComponent != "" {
			newParams.Set("activeComponent", activeComponent)
		}

		parsed.RawQuery = newParams.Encode()
	}

	parsed.Fragment = ""

	return parsed.String()
}

func validate(url string) bool {
	return url != "" && !strings.Contains(url, "fckdsh")
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

func crawl(rootUrlStr, startUrl string, maxParallelRequests, maxDepth int, client Client) int {

	var visited map[string]bool
	var queue chan Task
	var pool chan bool

	rootUrl, err := _url_.Parse(rootUrlStr)

	if err != nil {
		return 0
	}

	startUrl = normalize(rootUrl, startUrl)

	if !validate(startUrl) {
		return 0
	}

	visited = make(map[string]bool)
	queue = make(chan Task)
	pool = make(chan bool, maxParallelRequests)

	id := 1
	waiting := 1
	visited[startUrl] = true

	go request(id, startUrl, 0, client, queue, pool)

	for waiting > 0 {

		task := <-queue

		waiting--

		if task.urls != nil && (maxDepth == -1 || task.depth <= maxDepth) {

			for _, url := range task.urls {

				url = normalize(rootUrl, url)

				if validate(url) && !visited[url] {

					id++
					waiting++
					visited[url] = true

					go request(id, url, task.depth, client, queue, pool)

				}

			}
		}

	}

	return len(visited)

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
	maxParallelRequests := 10
	maxDepth := 2

	visited := crawl("http://localhost:8080", "/fx/auth", maxParallelRequests, maxDepth, bro.New(logLevel))

	log.Printf("Visited %d url(s)", visited)
}
