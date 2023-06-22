package main

import (
	"time"

	"github.com/freepaddler/yap-metrics/internal/agent/collector"
	"github.com/freepaddler/yap-metrics/internal/agent/config"
	"github.com/freepaddler/yap-metrics/internal/agent/reporter"
	"github.com/freepaddler/yap-metrics/internal/logger"
	"github.com/freepaddler/yap-metrics/internal/store/memory"
)

func main() {

	l := &logger.L

	conf := config.NewConfig()

	l.Info().Msg("Starting agent...")
	//return
	// collector should place data in storage
	// reported should report data from storage, set counters in storage as reported

	// new memory storage
	storage := memory.NewMemStorage()
	//rpt := reporter.NewPrintReporter(storage)
	rpt := reporter.NewHTTPReporter(storage, conf.ServerAddress, conf.HTTPTimeout)

	l.Debug().Msg("Starting loop")
	ticker := 0
	for {
		l.Debug().Msgf("ticker: %d", ticker)
		if ticker%int(conf.PollInterval) == 0 {
			collector.CollectMetrics(storage)
		}
		if ticker%int(conf.ReportInterval) == 0 {
			l.Debug().Msgf("\n======\nNew Report\n")
			rpt.Report()
		}
		time.Sleep(1 * time.Second)
		ticker++
	}

	// FIXME: this is never reachable until process control implementation
	//l.Info().Msg("Stopping agent...")
}
