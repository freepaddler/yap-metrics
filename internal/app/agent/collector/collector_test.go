package collector

import (
	"context"
	"fmt"
	"runtime"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/freepaddler/yap-metrics/mocks"
)

func TestCollector_CollectMetrics(t *testing.T) {
	var mockController = gomock.NewController(t)
	defer mockController.Finish()
	m := mocks.NewMockAgentController(mockController)
	c := New(m)
	m.EXPECT().CollectCounter("PollCount", int64(1)).Times(2)
	m.EXPECT().CollectGauge("RandomValue", gomock.Any()).Times(2)
	m.EXPECT().CollectGauge("Alloc", gomock.Any()).Times(2)
	m.EXPECT().CollectGauge("BuckHashSys", gomock.Any()).Times(2)
	m.EXPECT().CollectGauge("Frees", gomock.Any()).Times(2)
	m.EXPECT().CollectGauge("GCCPUFraction", gomock.Any()).Times(2)
	m.EXPECT().CollectGauge("GCSys", gomock.Any()).Times(2)
	m.EXPECT().CollectGauge("HeapAlloc", gomock.Any()).Times(2)
	m.EXPECT().CollectGauge("HeapIdle", gomock.Any()).Times(2)
	m.EXPECT().CollectGauge("HeapInuse", gomock.Any()).Times(2)
	m.EXPECT().CollectGauge("HeapObjects", gomock.Any()).Times(2)
	m.EXPECT().CollectGauge("HeapReleased", gomock.Any()).Times(2)
	m.EXPECT().CollectGauge("HeapSys", gomock.Any()).Times(2)
	m.EXPECT().CollectGauge("LastGC", gomock.Any()).Times(2)
	m.EXPECT().CollectGauge("Lookups", gomock.Any()).Times(2)
	m.EXPECT().CollectGauge("MCacheInuse", gomock.Any()).Times(2)
	m.EXPECT().CollectGauge("MCacheSys", gomock.Any()).Times(2)
	m.EXPECT().CollectGauge("MSpanInuse", gomock.Any()).Times(2)
	m.EXPECT().CollectGauge("MSpanSys", gomock.Any()).Times(2)
	m.EXPECT().CollectGauge("Mallocs", gomock.Any()).Times(2)
	m.EXPECT().CollectGauge("NextGC", gomock.Any()).Times(2)
	m.EXPECT().CollectGauge("NumForcedGC", gomock.Any()).Times(2)
	m.EXPECT().CollectGauge("NumGC", gomock.Any()).Times(2)
	m.EXPECT().CollectGauge("OtherSys", gomock.Any()).Times(2)
	m.EXPECT().CollectGauge("PauseTotalNs", gomock.Any()).Times(2)
	m.EXPECT().CollectGauge("StackInuse", gomock.Any()).Times(2)
	m.EXPECT().CollectGauge("StackSys", gomock.Any()).Times(2)
	m.EXPECT().CollectGauge("Sys", gomock.Any()).Times(2)
	m.EXPECT().CollectGauge("TotalAlloc", gomock.Any()).Times(2)
	c.CollectMetrics()
	c.CollectMetrics()
}

func TestCollector_CollectGOPSMetrics(t *testing.T) {
	var mockController = gomock.NewController(t)
	defer mockController.Finish()
	m := mocks.NewMockAgentController(mockController)
	c := New(m)
	m.EXPECT().CollectGauge("TotalMemory", gomock.Any()).Times(2)
	m.EXPECT().CollectGauge("FreeMemory", gomock.Any()).Times(2)
	for i := 0; i < runtime.NumCPU(); i++ {
		j := i + 1
		cpuMetric := fmt.Sprintf("CPUutilization%d", j)
		m.EXPECT().CollectGauge(cpuMetric, gomock.Any()).Times(2)
	}
	c.CollectGOPSMetrics(context.Background())
	c.CollectGOPSMetrics(context.Background())
}
