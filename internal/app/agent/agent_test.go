package agent

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/freepaddler/yap-metrics/internal/pkg/models"
	"github.com/freepaddler/yap-metrics/mocks"
)

func TestAgent_Run(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()
	store := mocks.NewMockAgentStorage(mockController)
	reporter := mocks.NewMockReporter(mockController)

	metrics := []models.Metrics{{Name: "name", Type: models.Counter, IValue: new(int64)}}
	tests := []struct {
		name           string
		runningTime    time.Duration
		reportInterval uint32
		pollInterval   uint32
		collectorRuns  int
		reporterRuns   int
		restoreCalls   int
		lastResult     error
		wantErr        error
	}{
		{
			name:           "normal stop",
			runningTime:    11 * time.Second,
			reportInterval: 2,
			pollInterval:   2,
			collectorRuns:  6,
			reporterRuns:   7,
			restoreCalls:   2,
		},
		{
			name:           "bad stop",
			runningTime:    11 * time.Second,
			reportInterval: 2,
			pollInterval:   2,
			collectorRuns:  6,
			reporterRuns:   7,
			restoreCalls:   2,
			lastResult:     errors.New("some err"),
			wantErr:        ErrPostShutdown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collectorRuns := 0
			collectorFunc := func(context.Context, CollectorStorage) {
				collectorRuns++
			}
			app := NewAgent(
				WithStore(store),
				WithCollectorFunc(collectorFunc),
				WithPollInterval(tt.pollInterval),
				WithReporter(reporter),
				WithReportInterval(tt.reportInterval),
				WithRateLimit(1),
			)
			ctx, cancel := context.WithTimeout(context.Background(), tt.runningTime)
			defer cancel()
			ts := time.Now()
			store.EXPECT().ReportAll().Times(tt.reporterRuns).Return(metrics, ts)
			//reporter.EXPECT().Send(metrics).Times(tt.reporterRuns)
			//store.EXPECT().RestoreLatest(metrics, ts).Times(tt.restoreCalls)
			gomock.InOrder(
				reporter.EXPECT().Send(metrics).Return(nil),
				reporter.EXPECT().Send(metrics).Return(nil),
				reporter.EXPECT().Send(metrics).Return(errors.New("some error")),
				store.EXPECT().RestoreLatest(metrics, ts),
				reporter.EXPECT().Send(metrics).Return(nil),
				reporter.EXPECT().Send(metrics).Return(errors.New("some error")),
				store.EXPECT().RestoreLatest(metrics, ts),
				reporter.EXPECT().Send(metrics).Return(nil),
				reporter.EXPECT().Send(metrics).Return(tt.lastResult),
			)
			err := app.Run(ctx)
			if tt.wantErr != nil {
				require.ErrorIsf(t, err, tt.wantErr, "Expect error '%s', got '%s'", tt.wantErr, err)
			}
			assert.Equal(t, tt.collectorRuns, collectorRuns, "Expect %d runs of collector, got %d", tt.collectorRuns, collectorRuns)
		})
	}

}
