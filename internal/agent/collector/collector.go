package collector

import (
	"fmt"
	"math/rand"
	"runtime"

	"github.com/freepaddler/yap-metrics/internal/store"
)

func CollectMetrics(s store.Storage) {
	fmt.Println("Start metrics collection routine")

	// update PollCount metric
	s.IncCounter("PollCount", 1)

	// update RandomValue
	s.SetGauge("RandomValue", rand.Float64())

	// update memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	//s.SetGauge("Alloc", float64(memStats.Alloc))
	//s.SetGauge("BuckHashSys", float64(memStats.BuckHashSys))
	//s.SetGauge("Frees", float64(memStats.Frees))
	//s.SetGauge("GCCPUFraction", memStats.GCCPUFraction)
	//s.SetGauge("GCSys", float64(memStats.GCSys))
	//s.SetGauge("HeapAlloc", float64(memStats.HeapAlloc))
	//s.SetGauge("HeapIdle", float64(memStats.HeapIdle))
	//s.SetGauge("HeapInuse", float64(memStats.HeapInuse))
	//s.SetGauge("HeapObjects", float64(memStats.HeapObjects))
	//s.SetGauge("HeapReleased", float64(memStats.HeapReleased))
	//s.SetGauge("HeapSys", float64(memStats.HeapSys))
	//s.SetGauge("LastGC", float64(memStats.LastGC))
	//s.SetGauge("Lookups", float64(memStats.Lookups))
	//s.SetGauge("MCacheInuse", float64(memStats.MCacheInuse))
	//s.SetGauge("MCacheSys", float64(memStats.MCacheSys))
	//s.SetGauge("MSpanInuse", float64(memStats.MSpanInuse))
	//s.SetGauge("MSpanSys", float64(memStats.MSpanSys))
	//s.SetGauge("Mallocs", float64(memStats.Mallocs))
	//s.SetGauge("NextGC", float64(memStats.NextGC))
	//s.SetGauge("NumForcedGC", float64(memStats.NumForcedGC))
	//s.SetGauge("NumGC", float64(memStats.NumGC))
	//s.SetGauge("OtherSys", float64(memStats.OtherSys))
	//s.SetGauge("PauseTotalNs", float64(memStats.PauseTotalNs))
	//s.SetGauge("StackInuse", float64(memStats.StackInuse))
	//s.SetGauge("StackSys", float64(memStats.StackSys))
	//s.SetGauge("Sys", float64(memStats.Sys))
	//s.SetGauge("TotalAlloc", float64(memStats.TotalAlloc))

	fmt.Println("Done metrics collection routine")
}
