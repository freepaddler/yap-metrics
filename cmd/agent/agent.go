package main

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/freepaddler/yap-metrics/internal/app/agent"
	"github.com/freepaddler/yap-metrics/internal/app/agent/collector"
	"github.com/freepaddler/yap-metrics/internal/app/agent/config"
	"github.com/freepaddler/yap-metrics/internal/app/agent/reporter/grpcbatchreporter"
	"github.com/freepaddler/yap-metrics/internal/app/agent/reporter/httpbatchreporter"
	"github.com/freepaddler/yap-metrics/internal/pkg/crypt"
	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
	"github.com/freepaddler/yap-metrics/internal/pkg/store"
	"github.com/freepaddler/yap-metrics/internal/pkg/store/memory"
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

	// try to load public key
	var pubKey *rsa.PublicKey

	if conf.PublicKeyFile != "" {
		pFile, err := os.Open(conf.PublicKeyFile)
		if err != nil {
			logger.Log().Error().Err(err).Msgf("unable to open public key file: %s", conf.PublicKeyFile)
			exitCode = 1
			return
		}
		pubKey, err = crypt.ReadPublicKey(pFile)
		if err != nil {
			logger.Log().Error().Err(err).Msgf("unable to read public key from file: %s", conf.PublicKeyFile)
			exitCode = 1
			return
		}
	}

	// notify context
	nCtx, nStop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer nStop()

	// start pprof webserver
	if conf.PprofAddress != "" {
		// pprof http server
		httpServer := &http.Server{
			Addr: conf.PprofAddress,
		}
		go func() {
			logger.Log().Info().Msgf("starting pprof http server at 'http://%s/debug/pprof'", httpServer.Addr)
			if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.Log().Error().Msg("failed to start pprof http server")
			}
			logger.Log().Info().Msg("pprof http server stopped acquiring new connections")
		}()
		defer func() {
			// context for httpServer graceful shutdown
			httpCtx, httpRelease := context.WithTimeout(context.Background(), 30*time.Second)
			defer httpRelease()

			// gracefully stop pprof http server
			logger.Log().Info().Msg("stopping pprof http server")
			if err := httpServer.Shutdown(httpCtx); err != nil {
				logger.Log().Err(err).Msg("failed to stop  pprof http server gracefully. force stop")
				_ = httpServer.Close()
			}
			logger.Log().Info().Msg("pprof http server stopped")
		}()
	}

	// setup reporter
	var reporter agent.Reporter
	if conf.GRPCServerAddress != "" {
		var err error
		reporter, err = grpcbatchreporter.New(
			grpcbatchreporter.WithAddress(conf.GRPCServerAddress),
			grpcbatchreporter.WithTimeout(conf.HTTPTimeout),
		)
		if err != nil {
			logger.Log().Error().Err(err).Msg("unable to init grpc reporter")
			exitCode = 2
			return
		}
	} else {
		reporter = httpbatchreporter.New(
			httpbatchreporter.WithAddress(conf.ServerAddress),
			httpbatchreporter.WithHTTPTimeout(conf.HTTPTimeout),
			httpbatchreporter.WithSignKey(conf.Key),
			httpbatchreporter.WithPublicKey(pubKey),
		)
	}

	// init and run agent
	app := agent.NewAgent(
		agent.WithStore(store.NewStorageController(memory.NewMemoryStore())),
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
		logger.Log().Error().Err(err).Msg("unclean exit")
		exitCode = 2
	}
	logger.Log().Info().Msg("agent stopped")
}
