package collector

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
	"github.com/freepaddler/yap-metrics/internal/pkg/models"
	"github.com/freepaddler/yap-metrics/internal/pkg/store"
)

type Collector struct {
	storage store.Storage
}

func New(s store.Storage) *Collector {
	return &Collector{storage: s}
}

func (c *Collector) collectCounter(n string, v int64) {
	c.storage.UpdateMetrics([]models.Metrics{{
		Name:   n,
		Type:   models.Counter,
		IValue: &v,
	}}, false)
}

func (c *Collector) collectGauge(n string, v float64) {
	c.storage.UpdateMetrics([]models.Metrics{{
		Name:   n,
		Type:   models.Gauge,
		FValue: &v,
	}}, false)
}

func (c *Collector) CollectMetrics() {
	logger.Log().Debug().Msg("start metrics collection routine")

	// update PollCount metric
	c.collectCounter("PollCount", 1)

	// update RandomValue
	c.collectGauge("RandomValue", rand.Float64())

	// update memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	c.collectGauge("Alloc", float64(memStats.Alloc))
	c.collectGauge("BuckHashSys", float64(memStats.BuckHashSys))
	c.collectGauge("Frees", float64(memStats.Frees))
	c.collectGauge("GCCPUFraction", memStats.GCCPUFraction)
	c.collectGauge("GCSys", float64(memStats.GCSys))
	c.collectGauge("HeapAlloc", float64(memStats.HeapAlloc))
	c.collectGauge("HeapIdle", float64(memStats.HeapIdle))
	c.collectGauge("HeapInuse", float64(memStats.HeapInuse))
	c.collectGauge("HeapObjects", float64(memStats.HeapObjects))
	c.collectGauge("HeapReleased", float64(memStats.HeapReleased))
	c.collectGauge("HeapSys", float64(memStats.HeapSys))
	c.collectGauge("LastGC", float64(memStats.LastGC))
	c.collectGauge("Lookups", float64(memStats.Lookups))
	c.collectGauge("MCacheInuse", float64(memStats.MCacheInuse))
	c.collectGauge("MCacheSys", float64(memStats.MCacheSys))
	c.collectGauge("MSpanInuse", float64(memStats.MSpanInuse))
	c.collectGauge("MSpanSys", float64(memStats.MSpanSys))
	c.collectGauge("Mallocs", float64(memStats.Mallocs))
	c.collectGauge("NextGC", float64(memStats.NextGC))
	c.collectGauge("NumForcedGC", float64(memStats.NumForcedGC))
	c.collectGauge("NumGC", float64(memStats.NumGC))
	c.collectGauge("OtherSys", float64(memStats.OtherSys))
	c.collectGauge("PauseTotalNs", float64(memStats.PauseTotalNs))
	c.collectGauge("StackInuse", float64(memStats.StackInuse))
	c.collectGauge("StackSys", float64(memStats.StackSys))
	c.collectGauge("Sys", float64(memStats.Sys))
	c.collectGauge("TotalAlloc", float64(memStats.TotalAlloc))

	logger.Log().Debug().Msg("done metrics collection routine")
}

func (c *Collector) CollectGOPSMetrics(ctx context.Context) {
	logger.Log().Debug().Msg("start gops metrics collection routine")

	vm, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		logger.Log().Warn().Msg("unable to get VirtualMemory metrics")
	} else {
		c.collectGauge("TotalMemory", float64(vm.Total))
		c.collectGauge("FreeMemory", float64(vm.Free))
	}

	cpuP, err := cpu.PercentWithContext(ctx, 0, true)
	if err != nil {
		logger.Log().Warn().Msg("unable to get CPUutilization metrics")
	} else {
		for i, v := range cpuP {
			c.collectGauge(fmt.Sprintf("CPUutilization%d", i+1), v)
		}
	}

	logger.Log().Debug().Msg("done gops metrics collection routine")
}
