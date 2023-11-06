package collector

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/freepaddler/yap-metrics/internal/app/agent/controller"
	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
)

type Collector struct {
	controller controller.AgentController
}

func New(ac controller.AgentController) *Collector {
	return &Collector{controller: ac}
}

func (c *Collector) CollectMetrics() {
	logger.Log().Debug().Msg("start metrics collection routine")

	// update PollCount metric
	c.controller.CollectCounter("PollCount", 1)

	// update RandomValue
	c.controller.CollectGauge("RandomValue", rand.Float64())

	// update memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	c.controller.CollectGauge("Alloc", float64(memStats.Alloc))
	c.controller.CollectGauge("BuckHashSys", float64(memStats.BuckHashSys))
	c.controller.CollectGauge("Frees", float64(memStats.Frees))
	c.controller.CollectGauge("GCCPUFraction", memStats.GCCPUFraction)
	c.controller.CollectGauge("GCSys", float64(memStats.GCSys))
	c.controller.CollectGauge("HeapAlloc", float64(memStats.HeapAlloc))
	c.controller.CollectGauge("HeapIdle", float64(memStats.HeapIdle))
	c.controller.CollectGauge("HeapInuse", float64(memStats.HeapInuse))
	c.controller.CollectGauge("HeapObjects", float64(memStats.HeapObjects))
	c.controller.CollectGauge("HeapReleased", float64(memStats.HeapReleased))
	c.controller.CollectGauge("HeapSys", float64(memStats.HeapSys))
	c.controller.CollectGauge("LastGC", float64(memStats.LastGC))
	c.controller.CollectGauge("Lookups", float64(memStats.Lookups))
	c.controller.CollectGauge("MCacheInuse", float64(memStats.MCacheInuse))
	c.controller.CollectGauge("MCacheSys", float64(memStats.MCacheSys))
	c.controller.CollectGauge("MSpanInuse", float64(memStats.MSpanInuse))
	c.controller.CollectGauge("MSpanSys", float64(memStats.MSpanSys))
	c.controller.CollectGauge("Mallocs", float64(memStats.Mallocs))
	c.controller.CollectGauge("NextGC", float64(memStats.NextGC))
	c.controller.CollectGauge("NumForcedGC", float64(memStats.NumForcedGC))
	c.controller.CollectGauge("NumGC", float64(memStats.NumGC))
	c.controller.CollectGauge("OtherSys", float64(memStats.OtherSys))
	c.controller.CollectGauge("PauseTotalNs", float64(memStats.PauseTotalNs))
	c.controller.CollectGauge("StackInuse", float64(memStats.StackInuse))
	c.controller.CollectGauge("StackSys", float64(memStats.StackSys))
	c.controller.CollectGauge("Sys", float64(memStats.Sys))
	c.controller.CollectGauge("TotalAlloc", float64(memStats.TotalAlloc))

	logger.Log().Debug().Msg("done metrics collection routine")
}

func (c *Collector) CollectGOPSMetrics(ctx context.Context) {
	logger.Log().Debug().Msg("start gops metrics collection routine")

	vm, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		logger.Log().Warn().Msg("unable to get VirtualMemory metrics")
	} else {
		c.controller.CollectGauge("TotalMemory", float64(vm.Total))
		c.controller.CollectGauge("FreeMemory", float64(vm.Free))
	}

	cpuP, err := cpu.PercentWithContext(ctx, 0, true)
	if err != nil {
		logger.Log().Warn().Msg("unable to get CPUutilization metrics")
	} else {
		for i, v := range cpuP {
			c.controller.CollectGauge(fmt.Sprintf("CPUutilization%d", i+1), v)
		}
	}

	logger.Log().Debug().Msg("done gops metrics collection routine")
}
