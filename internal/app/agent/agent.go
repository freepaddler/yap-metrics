package agent

import (
	"context"
	"errors"
	"net/http"
	_ "net/http/pprof"
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
	"github.com/freepaddler/yap-metrics/internal/pkg/wpool"
)

type Agent struct {
	conf      *config.Config
	storage   *memory.MemStorage
	reporter  *reporter.HTTPReporter
	collector *collector.Collector
}

func New(c *config.Config) *Agent {
	agt := Agent{conf: c}
	agt.storage = memory.NewMemStorage()
	agt.reporter = reporter.NewHTTPReporter(agt.storage, agt.conf.ServerAddress, agt.conf.HTTPTimeout, agt.conf.Key)
	agt.collector = collector.New(agt.storage)
	return &agt
}

func (agt *Agent) Run() {
	logger.Log.Info().Msg("starting agent")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	httpServer := &http.Server{
		Addr: "127.0.0.1:8091",
	}

	// trap os signals
	go func() {
		shutdownSig := make(chan os.Signal, 1)
		signal.Notify(shutdownSig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

		// shutdown routine
		sig := <-shutdownSig
		logger.Log.Info().Msgf("got '%v' signal. agent shutdown routine...", sig)
		cancel()

		// context for httpServer graceful shutdown
		httpCtx, httpRelease := context.WithTimeout(context.Background(), 10*time.Second)
		defer httpRelease()

		// gracefully stop http server
		logger.Log.Info().Msg("stopping pprof http server")
		if err := httpServer.Shutdown(httpCtx); err != nil {
			logger.Log.Err(err).Msg("failed to stop pprof http server gracefully. force stop")
			_ = httpServer.Close()
		}
		logger.Log.Info().Msg("pprof http server stopped")
	}()

	// start http server for profiling
	if agt.conf.PprofAddress != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			logger.Log.Info().Msgf("starting pprof http server at 'http://%s/debug/pprof'", agt.conf.PprofAddress)
			if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.Log.Error().Msg("failed to start pprof http server")
			}
			logger.Log.Info().Msg("pprof http server stopped acquiring new connections")
		}()
	}

	// start collection loops
	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()
		logger.Log.Debug().Msgf("starting metrics polling every %d seconds", agt.conf.PollInterval)
		for {
			agt.collector.CollectMetrics()
			select {
			case <-time.After(time.Duration(agt.conf.PollInterval) * time.Second):
			case <-ctx.Done():
				logger.Log.Debug().Msg("metrics polling cancelled")
				return
			}
		}
	}(ctx)
	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()
		logger.Log.Debug().Msgf("starting gops metrics polling every %d seconds", agt.conf.PollInterval)
		for {
			agt.collector.CollectGOPSMetrics(ctx)
			select {
			case <-time.After(time.Duration(agt.conf.PollInterval) * time.Second):
			case <-ctx.Done():
				logger.Log.Debug().Msg("gops metrics polling cancelled")
				return
			}
		}
	}(ctx)
	wg.Add(1)
	// start reporting loop
	go func(ctx context.Context) {
		defer wg.Done()
		wp := wpool.New(ctx, agt.conf.ReportRateLimit)
		wp.SetStopTimeout(5 * time.Second)
		logger.Log.Debug().Msgf("starting metrics reporting every %d seconds", agt.conf.ReportInterval)
		for {
			if err := wp.Task(func() { agt.reporter.ReportBatchJSON(ctx) }); err != nil {
				logger.Log.Warn().Err(err).Msg("unable to add reposting task to wpool")
			}
			select {
			case <-time.After(time.Duration(agt.conf.ReportInterval) * time.Second):
			case <-ctx.Done():
				logger.Log.Debug().Msg("metrics reporting cancelled")
				<-wp.Stopped
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

	// stop pprof http server

}
