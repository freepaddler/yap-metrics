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
	reporter := agent.NewPrintReporter()

	fmt.Println("Starting loop")
	ticker := 0
	for {
		fmt.Println("ticker:", ticker)
		if ticker%pollInterval == 0 {
			collectMetrics(sc)
		}
		if ticker%reportInterval == 0 {
			fmt.Println("\n\n======\nNew Report\n")
			reportMetrics(sc, reporter)
		}
		time.Sleep(1 * time.Second)
		ticker++
	}

	// FIXME: this is never reachable until process control implementation
	//fmt.Println("Stopping agent...")
}
