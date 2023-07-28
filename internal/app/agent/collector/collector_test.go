package collector

import (
	"context"
	"fmt"
	"testing"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/freepaddler/yap-metrics/internal/pkg/store/memory"
)

// check that value is random (differs between runs)
func Test_CollectMetrics_RandomValue(t *testing.T) {
	storage := memory.NewMemStorage()
	c := New(storage)
	c.CollectMetrics()
	val1, ok := storage.GetGauge("RandomValue")
	require.Equal(t, true, ok)
	c.CollectMetrics()
	val2, ok := storage.GetGauge("RandomValue")
	require.Equal(t, true, ok)
	assert.NotEqual(t, val1, val2)
}

// check that poll count increases every collection cycle
func Test_CollectMetrics_PollCount(t *testing.T) {
	storage := memory.NewMemStorage()
	c := New(storage)
	c.CollectMetrics()
	val1, ok := storage.GetCounter("PollCount")
	require.Equal(t, true, ok)
	assert.NotEqual(t, int64(0), val1)
	c.CollectMetrics()
	val2, ok := storage.GetCounter("PollCount")
	require.Equal(t, true, ok)
	assert.Equal(t, int64(1), *val2-*val1)
}

// check existence of all stats metrics
func Test_CollectMetrics_All_Metrics_Exist(t *testing.T) {
	counters := []string{"PollCount"}
	storage := memory.NewMemStorage()
	c := New(storage)
	c.CollectMetrics()
	c.CollectGOPSMetrics(context.TODO())
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
		"TotalMemory",
		"FreeMemory",
	}
	cpus, _ := cpu.Counts(true)
	for i := 0; i < cpus; i++ {
		gauges = append(gauges, fmt.Sprintf("CPUutilization%d", i+1))
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
