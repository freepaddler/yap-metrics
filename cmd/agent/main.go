package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/freepaddler/yap-metrics/internal/app/agent"
	"github.com/freepaddler/yap-metrics/internal/app/agent/config"
	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
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

	// print running config
	logger.Log().Info().Interface("config", conf).Msg("done config")

	// notify context
	nCtx, nStop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	defer nStop()

	// init and run agent
	app := agent.New(conf)
	app.Run(nCtx)
}
