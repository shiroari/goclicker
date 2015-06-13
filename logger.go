package main

import "log"

type Logger struct{}

func (self *Logger) On(status int, url string, foundUrls []string, foundErrors []string) {
	if foundErrors != nil {
		logPageError(url, status, foundErrors)
	}
}

func logPageError(url string, status int, errors []string) {
	for err := range errors {
		log.Printf("Error page %d :: %s :: %s\n", status, url, err)
	}
}
