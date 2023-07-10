package agent

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/freepaddler/yap-metrics/internal/app/agent/collector"
	"github.com/freepaddler/yap-metrics/internal/app/agent/config"
	"github.com/freepaddler/yap-metrics/internal/app/agent/reporter"
	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
	"github.com/freepaddler/yap-metrics/internal/pkg/store/memory"
)

type App struct {
	conf     *config.Config
	storage  *memory.MemStorage
	reporter *reporter.HTTPReporter
}

func New(c *config.Config) *App {
	agt := App{conf: c}
	agt.storage = memory.NewMemStorage()
	agt.reporter = reporter.NewHTTPReporter(agt.storage, agt.conf.ServerAddress, agt.conf.HTTPTimeout)
	return &agt
}

func (agt *App) Run() {
	logger.Log.Info().Msg("starting agent")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// start application loop
	go func(ctx context.Context) {
		logger.Log.Debug().Msgf("starting metrics polling every %d seconds", agt.conf.PollInterval)
		tPoll := time.NewTicker(time.Duration(agt.conf.PollInterval) * time.Second)
		defer tPoll.Stop()
		logger.Log.Debug().Msgf("starting metrics reporting every %d seconds", agt.conf.ReportInterval)
		tRpt := time.NewTicker(time.Duration(agt.conf.ReportInterval) * time.Second)
		defer tRpt.Stop()
		for {
			select {
			case <-ctx.Done():
				logger.Log.Debug().Msg("metrics polling stopped")
				return
			case <-tPoll.C:
				collector.CollectMetrics(agt.storage)
			case <-tRpt.C:
				agt.reporter.ReportJSON()
			}
		}
	}(ctx)

	// trap os signals
	shutdownSig := make(chan os.Signal, 1)
	signal.Notify(shutdownSig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// shutdown routine
	sig := <-shutdownSig
	logger.Log.Info().Msgf("got '%v' signal. agent shutdown routine...", sig)

	// post tasks
	defer func() {
		// gracefully stop all context-related goroutines
		cancel()

		// send all metrics to server
		logger.Log.Info().Msg("sending all metrics to server on exit")
		agt.reporter.ReportJSON()
		logger.Log.Info().Msg("agent stopped")
	}()

}
