package main

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v8"
	flag "github.com/spf13/pflag"

	"github.com/freepaddler/yap-metrics/internal/agent"
)

const (
	defaultPollInterval   = 2
	defaultReportInterval = 10
	defaultServerAddress  = "127.0.0.1:8080"
	defaultHttpTimeout    = 1
)

// global configuration
type config struct {
	PollInterval   uint32        `env:"POLL_INTERVAL"`
	ReportInterval uint32        `env:"REPORT_INTERVAL"`
	ServerAddress  string        `env:"ADDRESS"`
	HttpTimeout    time.Duration `env:"HTTP_TIMEOUT"`
}

func main() {
	conf := config{}

	// cmd params
	flag.StringVarP(
		&conf.ServerAddress,
		"serverAddress",
		"a",
		defaultServerAddress,
		"metrics collector server address HOST:PORT",
	)
	flag.Uint32VarP(
		&conf.ReportInterval,
		"reportInterval",
		"r",
		defaultReportInterval,
		"how often to send metrics to server (in seconds)",
	)
	flag.Uint32VarP(
		&conf.PollInterval,
		"pollInterval",
		"p",
		defaultPollInterval,
		"how often to collect metrics (in seconds)",
	)
	flag.DurationVarP(
		&conf.HttpTimeout,
		"httpTimeout",
		"t",
		defaultHttpTimeout,
		"http server response timeout (in seconds)",
	)
	flag.Parse()

	// env vars
	if err := env.Parse(&conf); err != nil {
		fmt.Println("Error while parsing ENV", err)
	}

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
