// Package agent implements agent's business logic.
//
// Runs collectors every pollInterval.
//
// Sends reports every reportInterval.
// In case of failed send restores unreported values back to store.
//
// Sending reports supports retries with predefined intervals.
//
// rateLimit limits the maximum number of simultaneous report processes.

package agent

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
	"github.com/freepaddler/yap-metrics/internal/pkg/models"
	"github.com/freepaddler/yap-metrics/pkg/retry"
	"github.com/freepaddler/yap-metrics/pkg/wpool"
)

var (
	ErrPostShutdown = errors.New("post-shutdown routine failed")
)

//go:generate mockgen -source $GOFILE -package=mocks -destination ../../../mocks/AgentStorage_mock.go

// CollectorStorage allows collectors to access store with required methods
type CollectorStorage interface {
	CollectCounter(name string, val int64)
	CollectGauge(name string, val float64)
}

// CollectorFunc defines function type for collectors
type CollectorFunc func(context.Context, CollectorStorage)

// ReporterStorage defines methods required for metrics reporting
type ReporterStorage interface {
	ReportAll() ([]models.Metrics, time.Time)
	RestoreLatest(metrics []models.Metrics, ts time.Time)
}

// Reporter is used to report metrics
type Reporter interface {
	Send([]models.Metrics) error
}

// AgentStorage interface for agent App
type AgentStorage interface {
	CollectorStorage
	ReporterStorage
}

// Agent is agent application config
type Agent struct {
	storage        AgentStorage
	collectors     []CollectorFunc
	pollInterval   time.Duration
	reporter       Reporter
	reportInterval time.Duration
	wg             sync.WaitGroup
	retries        []int
	rateLimit      int
}

// NewAgent is an Agent constructor
func NewAgent(options ...func(*Agent)) *Agent {
	agent := &Agent{
		pollInterval:   10 * time.Second,
		reportInterval: 10 * time.Second,
		rateLimit:      1,
	}
	for _, opt := range options {
		opt(agent)
	}
	return agent
}

// WithStore defines app storage
func WithStore(s AgentStorage) func(*Agent) {
	return func(agt *Agent) {
		agt.storage = s
	}
}

// WithCollectorFunc adds collector to the agent
func WithCollectorFunc(c CollectorFunc) func(*Agent) {
	return func(agt *Agent) {
		agt.collectors = append(agt.collectors, c)
	}
}

// WithPollInterval sets collecting interval
func WithPollInterval(interval uint32) func(*Agent) {
	return func(agt *Agent) {
		agt.pollInterval = time.Duration(interval) * time.Second
	}
}

// WithReporter adds Reporter to the agent
func WithReporter(r Reporter) func(*Agent) {
	return func(agt *Agent) {
		agt.reporter = r
	}
}

// WithReportInterval sets reporting interval
func WithReportInterval(interval uint32) func(*Agent) {
	return func(agt *Agent) {
		agt.reportInterval = time.Duration(interval) * time.Second
	}
}

// WithRetries sets report retries count
func WithRetries(r ...int) func(*Agent) {
	return func(agt *Agent) {
		agt.retries = append(agt.retries, r...)
	}
}

// WithRateLimit sets parallel sending processes count
func WithRateLimit(rl int) func(*Agent) {
	return func(agt *Agent) {
		if rl > 0 {
			agt.rateLimit = rl
			return
		}
		logger.Log().Warn().Msgf("Invalid rate limit '%d', use default %d", rl, agt.rateLimit)
	}
}

// checkCtxCancel validates context is not cancelled, exit fatal if it is
func checkCtxCancel(ctx context.Context) {
	if err := ctx.Err(); err != nil {
		logger.Log().Fatal().Err(err).Msg("application terminated")
	}
}

func (agt *Agent) Run(ctx context.Context) error {
	checkCtxCancel(ctx)
	logger.Log().Info().Msg("starting agent")

	// start collection loop
	agt.wg.Add(1)
	go func(ctx context.Context) {
		defer agt.wg.Done()
		logger.Log().Info().Msgf("starting metrics polling every %.f seconds", agt.pollInterval.Seconds())
		for {
			logger.Log().Info().Msg("new collect cycle")
			var wgCollector sync.WaitGroup
			wgCollector.Add(len(agt.collectors))
			for _, c := range agt.collectors {
				c := c
				go func() {
					defer wgCollector.Done()
					c(ctx, agt.storage)
				}()
			}
			wgCollector.Wait()
			select {
			case <-time.After(agt.pollInterval):
			case <-ctx.Done():
				logger.Log().Info().Msg("metrics polling terminated")
				return
			}
		}
	}(ctx)

	// start reporting loop
	agt.wg.Add(1)
	go func(ctx context.Context) {
		defer agt.wg.Done()
		logger.Log().Info().Msgf("starting metrics reporting every %.f seconds", agt.pollInterval.Seconds())
		wp := wpool.New(ctx, agt.rateLimit)
		for {
			logger.Log().Info().Msg("new report cycle")

			if err := wp.Task(func() {
				report, ts := agt.storage.ReportAll()

				err := retry.WithStrategy(ctx, func(context.Context) error {
					err := func() (err error) {
						return agt.reporter.Send(report)
					}()
					return err
				}, retry.IsNetErr, agt.retries...)

				if err != nil {
					logger.Log().Info().Msg("restore unsent report to store")
					agt.storage.RestoreLatest(report, ts)
				}
			}); err != nil {
				logger.Log().Warn().Err(err).Msg("unable to add reporting task to wpool")
			}

			select {
			case <-time.After(agt.reportInterval):
			case <-ctx.Done():
				logger.Log().Debug().Msg("metrics reporting cancelled")
				<-wp.Stop()
				return
			}
		}
	}(ctx)

	logger.Log().Info().Msg("agent started")

	// waiting for main context to be cancelled
	<-ctx.Done()
	logger.Log().Info().Msg("got stop request. stopping agent")

	//wait until tasks stopped
	agt.wg.Wait()
	logger.Log().Info().Msg("agent stopped. running post-shutdown routines")

	// post-shutdown tasks

	// report all metrics to server
	logger.Log().Info().Msg("report metrics to server")
	report, _ := agt.storage.ReportAll()
	err := agt.reporter.Send(report)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrPostShutdown, err)
	}
	logger.Log().Info().Msg("post-shutdown routines done")
	return nil
}
