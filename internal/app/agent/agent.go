package agent

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/freepaddler/yap-metrics/internal/app/agent/collector"
	"github.com/freepaddler/yap-metrics/internal/app/agent/config"
	"github.com/freepaddler/yap-metrics/internal/app/agent/reporter"
	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
	"github.com/freepaddler/yap-metrics/internal/pkg/store/memory"
)

type Agent struct {
	conf     *config.Config
	storage  *memory.MemStorage
	reporter *reporter.HTTPReporter
}

func New(c *config.Config) *Agent {
	agt := Agent{conf: c}
	agt.storage = memory.NewMemStorage()
	agt.reporter = reporter.NewHTTPReporter(agt.storage, agt.conf.ServerAddress, agt.conf.HTTPTimeout)
	return &agt
}

func (agt *Agent) Run() {
	logger.Log.Info().Msg("starting agent")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	// trap os signals
	go func() {
		shutdownSig := make(chan os.Signal, 1)
		signal.Notify(shutdownSig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

		// shutdown routine
		sig := <-shutdownSig
		logger.Log.Info().Msgf("got '%v' signal. agent shutdown routine...", sig)
		cancel()
	}()

	// start collection loop
	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()
		logger.Log.Debug().Msgf("starting metrics polling every %d seconds", agt.conf.PollInterval)
		for {
			collector.CollectMetrics(agt.storage)
			select {
			case <-time.After(time.Duration(agt.conf.PollInterval) * time.Second):
			case <-ctx.Done():
				logger.Log.Debug().Msg("metrics polling cancelled")
				return
			}
		}
	}(ctx)
	wg.Add(1)
	// start reporting loop
	go func(ctx context.Context) {
		defer wg.Done()
		logger.Log.Debug().Msgf("starting metrics reporting every %d seconds", agt.conf.ReportInterval)
		for {
			agt.reporter.ReportBatchJSON(ctx)
			select {
			case <-time.After(time.Duration(agt.conf.ReportInterval) * time.Second):
			case <-ctx.Done():
				logger.Log.Debug().Msg("metrics reporting cancelled")
				return
			}
		}
	}(ctx)

	// wait until tasks stopped
	wg.Wait()

	// shutdown tasks

	// send all metrics to server
	logger.Log.Info().Msg("sending all metrics to server on exit with 15 seconds timeout")
	ctxRep, ctxRepCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer ctxRepCancel()
	agt.reporter.ReportBatchJSON(ctxRep)
	logger.Log.Info().Msg("agent stopped")

}
