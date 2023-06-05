package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/freepaddler/yap-metrics/internal/agent"
	"github.com/freepaddler/yap-metrics/internal/models"
)

// mock reporter to check only one requested metric
type singleReporter struct {
	RepMetric models.Metrics
	MType     string
	MName     string
}

var _ agent.Reporter = (*singleReporter)(nil)

func newSingleReporter(mType, mName string) *singleReporter {
	return &singleReporter{
		MType:     mType,
		MName:     mName,
		RepMetric: models.Metrics{},
	}
}

func (r *singleReporter) Report(m models.Metrics) bool {
	if m.Type == r.MType && m.Name == r.MName {
		r.RepMetric = m
		return true
	}
	return false
}

// check that value is random (differs between runs)
func Test_collectMetrics_RandomValue(t *testing.T) {
	sc := agent.NewStatsCollector()
	sr := newSingleReporter(models.Gauge, "RandomValue")
	collectMetrics(sc)
	sc.ReportAll(sr)
	rand := sr.RepMetric.Gauge
	assert.Equal(t, models.Gauge, sr.RepMetric.Type)
	assert.Equal(t, "RandomValue", sr.RepMetric.Name)
	collectMetrics(sc)
	sc.ReportAll(sr)
	assert.NotEqual(t, rand, sr.RepMetric.Gauge)
}

// check that poll count increases every collection cycle
func Test_collectMetrics_PollCount(t *testing.T) {
	sc := agent.NewStatsCollector()
	sr := newSingleReporter(models.Counter, "PollCount")
	mAfterRun := models.Metrics{
		Name:      "PollCount",
		Type:      models.Counter,
		Increment: 1,
		Value:     1,
	}
	collectMetrics(sc)
	sc.ReportAll(sr)
	assert.Equal(t, mAfterRun, sr.RepMetric)
	// 3 consequent runs
	collectMetrics(sc)
	collectMetrics(sc)
	collectMetrics(sc)
	sc.ReportAll(sr)
	mAfterRun = models.Metrics{
		Name:      "PollCount",
		Type:      models.Counter,
		Increment: 3,
		Value:     4,
	}
	assert.Equal(t, mAfterRun, sr.RepMetric)
}

// check existence of all stats metrics
func Test_collectMetrics_exist(t *testing.T) {
	counters := []string{"PollCount"}
	gauges := []string{
		"RandomValue",
		"Alloc",
		"BuckHashSys",
		"Frees",
		"GCCPUFraction",
		"GCSys",
		"HeapAlloc",
		"HeapIdle",
		"HeapInuse",
		"HeapObjects",
		"HeapReleased",
		"HeapSys",
		"LastGC",
		"Lookups",
		"MCacheInuse",
		"MCacheSys",
		"MSpanInuse",
		"MSpanSys",
		"Mallocs",
		"NextGC",
		"NumForcedGC",
		"NumGC",
		"OtherSys",
		"PauseTotalNs",
		"StackInuse",
		"StackSys",
		"Sys",
		"TotalAlloc",
	}
	for _, v := range counters {
		sc := agent.NewStatsCollector()
		sr := newSingleReporter(models.Counter, v)
		collectMetrics(sc)
		sc.ReportAll(sr)
		assert.Equalf(t, models.Counter, sr.RepMetric.Type, "Counter %s type mismatch", v)
		assert.Equalf(t, v, sr.RepMetric.Name, "Counter %s name mismatch", v)
	}
	for _, v := range gauges {
		sc := agent.NewStatsCollector()
		sr := newSingleReporter(models.Gauge, v)
		collectMetrics(sc)
		sc.ReportAll(sr)
		assert.Equalf(t, models.Gauge, sr.RepMetric.Type, "Gauge %s type mismatch", v)
		assert.Equalf(t, v, sr.RepMetric.Name, "Gauge %s name mismatch", v)
	}

}
