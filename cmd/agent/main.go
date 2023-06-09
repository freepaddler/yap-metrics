package main

import (
	"fmt"
	"time"

	"github.com/freepaddler/yap-metrics/internal/agent"
	"github.com/freepaddler/yap-metrics/internal/agent/config"
)

func main() {
	conf := config.NewConfig()

	fmt.Printf(`Starting agent...
		server: %s
		pollInterval: %ds
		reportInterval: %ds
		httpTimeout: %ds
`, conf.ServerAddress, conf.PollInterval, conf.ReportInterval, conf.HttpTimeout)

	sc := agent.NewStatsCollector()
	reporter := agent.NewHTTPReporter(conf.ServerAddress, conf.HttpTimeout)
	//reporter := agent.NewPrintReporter()

	fmt.Println("Starting loop")
	ticker := 0
	for {
		fmt.Println("ticker:", ticker)
		if ticker%int(conf.PollInterval) == 0 {
			collectMetrics(sc)
		}
		if ticker%int(conf.ReportInterval) == 0 {
			fmt.Printf("\n\n======\nNew ReportAll\n\n")
			//sc.ReportAll(printReporter)
			sc.ReportAll(reporter)
			//reportMetrics(sc, &reporter)
		}
		time.Sleep(1 * time.Second)
		ticker++
	}

	// FIXME: this is never reachable until process control implementation
	//fmt.Println("Stopping agent...")
}
