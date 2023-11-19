package controller

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/freepaddler/yap-metrics/internal/pkg/models"
	"github.com/freepaddler/yap-metrics/mocks"
)

func TestMetricsController_CollectCounter(t *testing.T) {
	var mockController = gomock.NewController(t)
	defer mockController.Finish()
	m := mocks.NewMockMemoryStore(mockController)

	name := "counter1"
	value := int64(-12)

	c := New(m)

	m.EXPECT().IncCounter(name, value).Times(1)
	c.CollectCounter(name, value)
}

func TestMetricsController_CollectGauge(t *testing.T) {
	var mockController = gomock.NewController(t)
	defer mockController.Finish()
	m := mocks.NewMockMemoryStore(mockController)

	name := "gauge1"
	value := -0.117

	c := New(m)

	m.EXPECT().SetGauge(name, value).Times(2)
	tStart := time.Now()
	c.CollectGauge(name, value)
	require.WithinRange(t, c.gaugesTS[name], tStart, time.Now(), "Invalid timestamp in map")
	tStart = time.Now()
	c.CollectGauge(name, value)
	require.WithinRange(t, c.gaugesTS[name], tStart, time.Now(), "Invalid timestamp in map")
}

func TestMetricsController_ReportAll(t *testing.T) {
	var mockController = gomock.NewController(t)
	defer mockController.Finish()
	m := mocks.NewMockMemoryStore(mockController)

	c := New(m)

	m.EXPECT().Snapshot(true).Times(1)
	tStart := time.Now()
	_, ts := c.ReportAll()
	require.WithinRange(t, ts, tStart, time.Now(), "Unexpected report timestamp")
}

func TestMetricsController_RestoreReport(t *testing.T) {
	var mockController = gomock.NewController(t)
	defer mockController.Finish()
	m := mocks.NewMockMemoryStore(mockController)

	// counter in report
	cVal := int64(12)
	counter := models.Metrics{
		Name:   "c1",
		Type:   models.Counter,
		IValue: &cVal,
	}
	// gauge in report
	gVal := 1.119
	gauge := models.Metrics{
		Name:   "g1",
		Type:   models.Gauge,
		FValue: &gVal,
	}
	// report
	report := []models.Metrics{counter, gauge}

	t.Run("Restore to empty store", func(t *testing.T) {
		c := New(m)
		ts := time.Now().Add(-1 * time.Second)
		// both metrics should be updated
		m.EXPECT().IncCounter(counter.Name, *counter.IValue).Times(1)
		m.EXPECT().SetGauge(gauge.Name, *gauge.FValue).Times(1)
		c.RestoreReport(report, ts)
		require.Equal(t, ts, c.gaugesTS[gauge.Name], "Expect '%t' in gaugesTs map, got '%t'", ts, c.gaugesTS[gauge.Name])
	})

	t.Run("Restore to updated store", func(t *testing.T) {
		c := New(m)

		m.EXPECT().Snapshot(true).Return(report).Times(1)
		r, reportTS := c.ReportAll()

		m.EXPECT().IncCounter(gomock.Any(), gomock.Any())
		m.EXPECT().SetGauge(gomock.Any(), gomock.Any())
		c.CollectGauge("g1", 1)
		c.CollectCounter("c1", 1)

		// only counter should be updated
		m.EXPECT().IncCounter(counter.Name, *counter.IValue).Times(1)
		c.RestoreReport(r, reportTS)

		// gauge timestamp should not be changed
		require.True(t, c.gaugesTS[gauge.Name].After(reportTS), "Expect time in gaugesTs map later than report time")
	})
}
