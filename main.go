/**
 *
 *
 *
 **/

package main

import (
 	"log"
 	"net/http"
	"net/http/cookiejar"
	"client"
	"strings"
	"time"
)

type Fetcher interface {
	fetch(url string) (title string, urls []string, err error)
}

type Task struct {
	urls  []string
	depth int
}

func fetch(url string) (string, []string, error) {
	urls, err := client.GetUrls(client_, url)
	if err != nil {
		return "", nil, err
	}
	return "", urls, nil
}

func request(id int, url string, depth int, queue chan Task) {

	body, urls, err := fetch(url)

	if err != nil {
		log.Printf("[%d] error - %s\n", id, err)
		queue <- Task{nil, depth + 1}
		return
	}

	log.Printf("[%d] %s, title: %q, depth: %d, more: %d\n", id, url, body, depth, len(urls))

	queue <- Task{urls, depth + 1}

}

func crawl(root string, startUrl string, maxDepth int) {

	//var visited map[string]bool
	var queue chan Task

	visited = make(map[string]bool)
	queue = make(chan Task)

	id := 1
	waiting := 1
	visited[startUrl] = true

	go request(id, root + startUrl, 0, queue)

	for waiting > 0 {

		task := <-queue

		waiting--

		if task.urls != nil && task.depth <= maxDepth {

			for _, url := range task.urls {

				if !visited[url] && !strings.Contains(url, "fckdsh") {

					id++
					waiting++
					visited[url] = true

					go request(id, root + url, task.depth, queue)

				}

			}
		}

	}

}

var visited map[string]bool
var client_ *http.Client

func timeTrack(start time.Time, name string) {
    elapsed := time.Since(start)
    log.Printf("%s took %s", name, elapsed)
}

func main() {

	defer timeTrack(time.Now(), "App")

	client_ = &http.Client{}
	jar, err := cookiejar.New(nil)

	if err != nil {
		log.Fatal(err)
	}

    client_.Jar = jar

	crawl("http://localhost:8080" , "/fx/auth", 2)

	log.Printf("Visited: %d", len(visited))

}
