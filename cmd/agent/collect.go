package main

import (
	"fmt"
	"math/rand"
	"runtime"

	"github.com/freepaddler/yap-metrics/internal/agent"
)

func collectMetrics(sc *agent.StatsCollector) {
	fmt.Println("Start metrics collection routine")

	// update PollCount metric
	sc.Counter("PollCount").Inc(1)

	// update RandomValue
	rValue := rand.Float64()
	sc.Gauge("RandomValue").Update(rValue)

	// update memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	sc.Gauge("Alloc").Update(float64(memStats.Alloc))
	sc.Gauge("BuckHashSys").Update(float64(memStats.BuckHashSys))
	sc.Gauge("Frees").Update(float64(memStats.Frees))
	sc.Gauge("GCCPUFraction").Update(memStats.GCCPUFraction)
	sc.Gauge("GCSys").Update(float64(memStats.GCSys))
	sc.Gauge("HeapAlloc").Update(float64(memStats.HeapAlloc))
	sc.Gauge("HeapIdle").Update(float64(memStats.HeapIdle))
	sc.Gauge("HeapInuse").Update(float64(memStats.HeapInuse))
	sc.Gauge("HeapObjects").Update(float64(memStats.HeapObjects))
	sc.Gauge("HeapReleased").Update(float64(memStats.HeapReleased))
	sc.Gauge("HeapSys").Update(float64(memStats.HeapSys))
	sc.Gauge("LastGC").Update(float64(memStats.LastGC))
	sc.Gauge("Lookups").Update(float64(memStats.Lookups))
	sc.Gauge("MCacheInuse").Update(float64(memStats.MCacheInuse))
	sc.Gauge("MCacheSys").Update(float64(memStats.MCacheSys))
	sc.Gauge("MSpanInuse").Update(float64(memStats.MSpanInuse))
	sc.Gauge("MSpanSys").Update(float64(memStats.MSpanSys))
	sc.Gauge("Mallocs").Update(float64(memStats.Mallocs))
	sc.Gauge("NextGC").Update(float64(memStats.NextGC))
	sc.Gauge("NumForcedGC").Update(float64(memStats.NumForcedGC))
	sc.Gauge("NumGC").Update(float64(memStats.NumGC))
	sc.Gauge("OtherSys").Update(float64(memStats.OtherSys))
	sc.Gauge("PauseTotalNs").Update(float64(memStats.PauseTotalNs))
	sc.Gauge("StackInuse").Update(float64(memStats.StackInuse))
	sc.Gauge("StackSys").Update(float64(memStats.StackSys))
	sc.Gauge("Sys").Update(float64(memStats.Sys))
	sc.Gauge("TotalAlloc").Update(float64(memStats.TotalAlloc))

	fmt.Println("Done metrics collection routine")
}
