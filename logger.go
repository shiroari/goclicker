package main

import (
	"log"
	net "net/url"
)

type Logger struct{}

var prefixes = map[string]int{}

func (self *Logger) On(status int, url string, foundUrls []string, foundErrors []string) {
	if foundErrors != nil {
		logPageError(url, status, foundErrors)
	}
	parsedUrl, err := net.Parse(url)
	if err != nil {
		log.Panicln(err)
		return
	}
	uuid := parsedUrl.Query().Get("uuid")
	if len(uuid) > 6 {
		prefix := uuid[:6]
		prefixes[prefix]++
	}
}

func (self *Logger) GetStat() map[string]int {
	return prefixes
}

func (self *Logger) PrintStat() {
	log.Print("Prefix\tVisited\n")
	total := 0
	for key, value := range prefixes {
		total += value
		log.Printf("%s\t%d\n", key, value)
	}
	log.Printf("Total\t%d\n", total)
}

func logPageError(url string, status int, errors []string) {
	for err := range errors {
		log.Printf("Error page %d :: %s :: %s\n", status, url, err)
	}
}
