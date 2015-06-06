/**
 * Simple Crawler Tool
 *
 * TODO:
 * 	  - remove link starting with javascript:,mailto:,sip:
 *	  - collect links inside inline javascript (onclick)
 * 	  - find error pages
 *	  - sortout url parameters
 *    - limit requests
 *    - detect long waiting pages
 **/

package main

import (
	"log"
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

func normalize(host, url string) string {
	if url == "#" {
		return ""
	}
	return host + url
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

func request(id int, url string, depth int, client Client, queue chan Task) {

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

func crawl(host string, startUrl string, maxDepth int, client Client) int {

	var visited map[string]bool
	var queue chan Task

	visited = make(map[string]bool)
	queue = make(chan Task)

	startUrl = normalize(host, startUrl)

	if !validate(startUrl) {
		return 0
	}

	id := 1
	waiting := 1
	visited[startUrl] = true

	go request(id, startUrl, 0, client, queue)

	for waiting > 0 {

		task := <-queue

		waiting--

		if task.urls != nil && (maxDepth == -1 || task.depth <= maxDepth) {

			for _, url := range task.urls {

				url = normalize(host, url)

				if validate(url) && !visited[url] {

					id++
					waiting++
					visited[url] = true

					go request(id, url, task.depth, client, queue)

				}

			}
		}

	}

	return len(visited)

}

var logLevel int

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s time: %s", name, elapsed)
}

func main() {

	runtime.GOMAXPROCS(2)

	var mem1, mem2 runtime.MemStats

	runtime.ReadMemStats(&mem1)

	logLevel = 0

	start := time.Now()

	client := bro.New(logLevel)

	visited := crawl("http://localhost:8080", "/fx/auth", 5, client)

	log.Printf("Visited: %d url(s) for %s", visited, time.Since(start))

	runtime.GC()
	runtime.ReadMemStats(&mem2)

	log.Printf("Memory stat: %.2f Mb\n", float64(mem2.Sys-mem1.Sys)/(1024*1024))
	log.Printf("Number of Goroutines: %d\n", runtime.NumGoroutine())
}
