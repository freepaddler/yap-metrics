package collector

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/freepaddler/yap-metrics/internal/store/memory"
)

// mock reporter to check only one requested metric
//type singleReporter struct {
//	RepMetric models.Metrics
//	MType     string
//	MName     string
//}

//func newSingleReporter(mType, mName string) *singleReporter {
//	return &singleReporter{
//		MType:     mType,
//		MName:     mName,
//		RepMetric: models.Metrics{},
//	}
//}

//func (r *singleReporter) Report(m models.Metrics) bool {
//	if m.Type == r.MType && m.Name == r.MName {
//		r.RepMetric = m
//		return true
//	}
//	return false
//}

// check that value is random (differs between runs)
func Test_CollectMetrics_RandomValue(t *testing.T) {
	storage := memory.NewMemStorage()
	CollectMetrics(storage)
	val1, ok := storage.GetGauge("RandomValue")
	require.Equal(t, true, ok)
	CollectMetrics(storage)
	val2, ok := storage.GetGauge("RandomValue")
	require.Equal(t, true, ok)
	assert.NotEqual(t, val1, val2)
}

// check that poll count increases every collection cycle
func Test_CollectMetrics_PollCount(t *testing.T) {
	storage := memory.NewMemStorage()
	CollectMetrics(storage)
	val1, ok := storage.GetCounter("PollCount")
	require.Equal(t, true, ok)
	assert.NotEqual(t, int64(0), val1)
	CollectMetrics(storage)
	val2, ok := storage.GetCounter("PollCount")
	require.Equal(t, true, ok)
	assert.Equal(t, int64(1), *val2-*val1)
}

// check existence of all stats metrics
func Test_CollectMetrics_All_Metrics_Exist(t *testing.T) {
	counters := []string{"PollCount"}
	storage := memory.NewMemStorage()
	CollectMetrics(storage)
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
		_, ok := storage.GetCounter(v)
		require.Equalf(t, true, ok, "Missing counter %s", v)
	}
	for _, v := range gauges {
		_, ok := storage.GetGauge(v)
		require.Equalf(t, true, ok, "Missing gauge %s", v)
	}

}
