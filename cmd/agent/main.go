package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/freepaddler/yap-metrics/internal/agent"
)

const (
	pollInterval   = 2
	reportInterval = 10
	address        = "127.0.0.1:8080"
)

var (
	// random generator source
	r *rand.Rand
)

func init() {
	// init random generator source
	r = rand.New(rand.NewSource(rand.Int63()))
}

func main() {
	fmt.Println("Starting agent...")

	sc := agent.NewStatsCollector()
	//var reporter agent.Reporter
	//printReporter := agent.NewPrintReporter()
	httpReporter := agent.NewHttpReporter(address)

	fmt.Println("Starting loop")
	ticker := 0
	for {
		fmt.Println("ticker:", ticker)
		if ticker%pollInterval == 0 {
			collectMetrics(sc)
		}
		if ticker%reportInterval == 0 {
			fmt.Printf("\n\n======\nNew Report\n\n")
			//sc.Report(printReporter)
			sc.Report(httpReporter)
			//reportMetrics(sc, &reporter)
		}
		time.Sleep(1 * time.Second)
		ticker++
	}

	// FIXME: this is never reachable until process control implementation
	//fmt.Println("Stopping agent...")
}
