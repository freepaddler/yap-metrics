package main

import (
	"fmt"
	"math/rand"
	"time"

	flag "github.com/spf13/pflag"

	"github.com/freepaddler/yap-metrics/internal/agent"
)

const (
	defaultPollInterval   = 2
	defaultReportInterval = 10
	defaultServerAddress  = "127.0.0.1:8080"
)

var (
	// random generator source
	r *rand.Rand
)

type config struct {
	pollInterval   uint32
	reportInterval uint32
	serverAddress  string
}

func init() {
	// init random generator source
	r = rand.New(rand.NewSource(rand.Int63()))
}

func main() {
	conf := config{}
	flag.StringVarP(
		&conf.serverAddress,
		"serverAddress",
		"a",
		defaultServerAddress,
		"metrics collector server address HOST:PORT",
	)
	flag.Uint32VarP(
		&conf.reportInterval,
		"reportInterval",
		"r",
		defaultReportInterval,
		"how often to send metrics to server (in seconds)",
	)
	flag.Uint32VarP(
		&conf.pollInterval,
		"pollInterval",
		"p",
		defaultPollInterval,
		"how often to collect metrics (in seconds)",
	)
	flag.Parse()

	fmt.Printf(`Starting agent...
		server: %s
		pollInterval: %d seconds
		reportInterval: %d seconds
`, conf.serverAddress, conf.pollInterval, conf.reportInterval)

	sc := agent.NewStatsCollector()
	//var reporter agent.Reporter
	//printReporter := agent.NewPrintReporter()
	httpReporter := agent.NewHTTPReporter(conf.serverAddress)

	fmt.Println("Starting loop")
	ticker := 0
	for {
		fmt.Println("ticker:", ticker)
		if ticker%int(conf.pollInterval) == 0 {
			collectMetrics(sc)
		}
		if ticker%int(conf.reportInterval) == 0 {
			fmt.Printf("\n\n======\nNew ReportAll\n\n")
			//sc.ReportAll(printReporter)
			sc.ReportAll(httpReporter)
			//reportMetrics(sc, &reporter)
		}
		time.Sleep(1 * time.Second)
		ticker++
	}

	// FIXME: this is never reachable until process control implementation
	//fmt.Println("Stopping agent...")
}
