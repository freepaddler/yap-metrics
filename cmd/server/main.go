package main

import (
	"fmt"
	_ "net/http/pprof"
	"os"

	"github.com/freepaddler/yap-metrics/internal/app/server"
	"github.com/freepaddler/yap-metrics/internal/app/server/config"
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

	// server configuration
	conf := config.NewConfig()

	// set log level
	logger.SetLevel(conf.LogLevel)

	// print running config
	logger.Log().Info().Interface("config", conf).Msg("done config")

	// init and run server
	app := server.New(conf)
	app.Run()
}
