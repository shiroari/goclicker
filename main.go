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
	"runtime"
	"time"

	"robot/bro"
	c "robot/crawler"
)

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
	logger := &Logger{}

	crawler := c.Crawler{
		"http://localhost:8080",
		maxParallelRequests,
		maxDepth,
		bro.New("system", "manager", logLevel),
		&Parser{},
		logger,
		logLevel}

	logger.PrintStat()

	visited := crawler.Start("/fx/auth")

	log.Printf("Visited %d url(s)", visited)
}
