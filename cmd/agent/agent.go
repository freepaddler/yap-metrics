package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/freepaddler/yap-metrics/internal/app/agent"
	"github.com/freepaddler/yap-metrics/internal/app/agent/collector"
	"github.com/freepaddler/yap-metrics/internal/app/agent/config"
	"github.com/freepaddler/yap-metrics/internal/app/agent/controller"
	"github.com/freepaddler/yap-metrics/internal/app/agent/reporter/httpBatchReporter"
	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
	"github.com/freepaddler/yap-metrics/internal/pkg/store/inmemory"
)

var (
	// go build -ldflags " \
	// -X 'main.buildVersion=$(git describe --tag 2>/dev/null)' \
	// -X 'main.buildDate=$(date)' \
	// -X 'main.buildCommit=$(git rev-parse --short HEAD)' \
	// "
	buildVersion, buildDate, buildCommit = "N/A", "N/A", "N/A"
)

func main() {
	exitCode := 0
	defer func() { os.Exit(exitCode) }()

	fmt.Fprintf(
		os.Stdout,
		`Build version: %s
Build date: %s
Build commit %s
`, buildVersion, buildDate, buildCommit)

	// agent configuration
	conf := config.NewConfig()

	// set log level
	logger.SetLevel(conf.LogLevel)

	// notify context
	nCtx, nStop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer nStop()

	// setup reporter
	reporter := httpBatchReporter.New(
		httpBatchReporter.WithAddress(conf.ServerAddress),
		httpBatchReporter.WithHTTPTimeout(conf.HTTPTimeout),
		httpBatchReporter.WithSignKey(conf.Key),
		httpBatchReporter.WithPublicKey(conf.PublicKey),
	)

	// init and run agent
	app := agent.New(
		agent.WithStore(controller.New(inmemory.New())),
		agent.WithCollectorFunc(collector.Simple),
		agent.WithCollectorFunc(collector.MemStats),
		agent.WithCollectorFunc(collector.GoPS),
		agent.WithPollInterval(conf.PollInterval),
		agent.WithReporter(reporter),
		agent.WithReportInterval(conf.ReportInterval),
		agent.WithRetries(1, 3, 5),
		agent.WithRateLimit(2),
	)
	err := app.Run(nCtx)
	if err != nil {
		exitCode = 2
	}
}
