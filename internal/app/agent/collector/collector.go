package collector

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/freepaddler/yap-metrics/internal/app/agent"
	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
)

func Simple(_ context.Context, store agent.StoreCollector) {
	logger.Log().Debug().Msg("collect simple start")
	// update PollCount metric
	store.CollectCounter("PollCount", 1)
	// update RandomValue
	store.CollectGauge("RandomValue", rand.Float64())
	logger.Log().Debug().Msg("collect simple done")
}

func MemStats(_ context.Context, store agent.StoreCollector) {
	logger.Log().Debug().Msg("collect memStats start")
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	store.CollectGauge("Alloc", float64(memStats.Alloc))
	store.CollectGauge("BuckHashSys", float64(memStats.BuckHashSys))
	store.CollectGauge("Frees", float64(memStats.Frees))
	store.CollectGauge("GCCPUFraction", memStats.GCCPUFraction)
	store.CollectGauge("GCSys", float64(memStats.GCSys))
	store.CollectGauge("HeapAlloc", float64(memStats.HeapAlloc))
	store.CollectGauge("HeapIdle", float64(memStats.HeapIdle))
	store.CollectGauge("HeapInuse", float64(memStats.HeapInuse))
	store.CollectGauge("HeapObjects", float64(memStats.HeapObjects))
	store.CollectGauge("HeapReleased", float64(memStats.HeapReleased))
	store.CollectGauge("HeapSys", float64(memStats.HeapSys))
	store.CollectGauge("LastGC", float64(memStats.LastGC))
	store.CollectGauge("Lookups", float64(memStats.Lookups))
	store.CollectGauge("MCacheInuse", float64(memStats.MCacheInuse))
	store.CollectGauge("MCacheSys", float64(memStats.MCacheSys))
	store.CollectGauge("MSpanInuse", float64(memStats.MSpanInuse))
	store.CollectGauge("MSpanSys", float64(memStats.MSpanSys))
	store.CollectGauge("Mallocs", float64(memStats.Mallocs))
	store.CollectGauge("NextGC", float64(memStats.NextGC))
	store.CollectGauge("NumForcedGC", float64(memStats.NumForcedGC))
	store.CollectGauge("NumGC", float64(memStats.NumGC))
	store.CollectGauge("OtherSys", float64(memStats.OtherSys))
	store.CollectGauge("PauseTotalNs", float64(memStats.PauseTotalNs))
	store.CollectGauge("StackInuse", float64(memStats.StackInuse))
	store.CollectGauge("StackSys", float64(memStats.StackSys))
	store.CollectGauge("Sys", float64(memStats.Sys))
	store.CollectGauge("TotalAlloc", float64(memStats.TotalAlloc))

	logger.Log().Debug().Msg("collect memStats done")
}

func GoPS(ctx context.Context, store agent.StoreCollector) {
	logger.Log().Debug().Msg("collect GoPS start")

	vm, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		logger.Log().Warn().Msg("unable to get VirtualMemory metrics")
	} else {
		store.CollectGauge("TotalMemory", float64(vm.Total))
		store.CollectGauge("FreeMemory", float64(vm.Free))
	}

	cpuP, err := cpu.PercentWithContext(ctx, 0, true)
	if err != nil {
		logger.Log().Warn().Msg("unable to get CPUutilization metrics")
	} else {
		for i, v := range cpuP {
			store.CollectGauge(fmt.Sprintf("CPUutilization%d", i+1), v)
		}
	}

	logger.Log().Debug().Msg("collect GoPS done")
}
