package main

import (
	"fmt"
	"time"

	"github.com/freepaddler/yap-metrics/internal/agent/collector"
	"github.com/freepaddler/yap-metrics/internal/agent/config"
	"github.com/freepaddler/yap-metrics/internal/agent/reporter"
	"github.com/freepaddler/yap-metrics/internal/store/memory"
)

func main() {
	conf := config.NewConfig()

	fmt.Printf(`Starting agent...
		server: %s
		pollInterval: %ds
		reportInterval: %ds
		httpTimeout: %ds
`, conf.ServerAddress, conf.PollInterval, conf.ReportInterval, conf.HttpTimeout)

	// collector should place data in storage
	// reported should report data from storage, set counters in storage as reported

	// new memory storage
	storage := memory.NewMemStorage()
	//rpt := reporter.NewPrintReporter(storage)
	rpt := reporter.NewHTTPReporter(storage, conf.ServerAddress, conf.HttpTimeout)

	fmt.Println("Starting loop")
	ticker := 0
	for {
		fmt.Println("ticker:", ticker)
		if ticker%int(conf.PollInterval) == 0 {
			collector.CollectMetrics(storage)
		}
		if ticker%int(conf.ReportInterval) == 0 {
			fmt.Printf("\n\n======\nNew Report\n\n")
			rpt.Report()
		}
		time.Sleep(1 * time.Second)
		ticker++
	}

	// FIXME: this is never reachable until process control implementation
	//fmt.Println("Stopping agent...")
}
