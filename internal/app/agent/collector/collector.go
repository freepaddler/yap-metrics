package collector

import (
	"math/rand"
	"runtime"

	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
	"github.com/freepaddler/yap-metrics/internal/pkg/models"
	"github.com/freepaddler/yap-metrics/internal/pkg/store"
)

func CollectMetrics(s store.Storage) {
	logger.Log.Debug().Msg("start metrics collection routine")

	collectCounter := func(n string, v int64) {
		s.UpdateMetrics([]models.Metrics{{
			Name:   n,
			Type:   models.Counter,
			IValue: &v,
		}}, false)
	}

	collectGauge := func(n string, v float64) {
		s.UpdateMetrics([]models.Metrics{{
			Name:   n,
			Type:   models.Gauge,
			FValue: &v,
		}}, false)
	}

	// update PollCount metric
	collectCounter("PollCount", 1)

	// update RandomValue
	collectGauge("RandomValue", rand.Float64())

	// update memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	collectGauge("Alloc", float64(memStats.Alloc))
	collectGauge("BuckHashSys", float64(memStats.BuckHashSys))
	collectGauge("Frees", float64(memStats.Frees))
	collectGauge("GCCPUFraction", memStats.GCCPUFraction)
	collectGauge("GCSys", float64(memStats.GCSys))
	collectGauge("HeapAlloc", float64(memStats.HeapAlloc))
	collectGauge("HeapIdle", float64(memStats.HeapIdle))
	collectGauge("HeapInuse", float64(memStats.HeapInuse))
	collectGauge("HeapObjects", float64(memStats.HeapObjects))
	collectGauge("HeapReleased", float64(memStats.HeapReleased))
	collectGauge("HeapSys", float64(memStats.HeapSys))
	collectGauge("LastGC", float64(memStats.LastGC))
	collectGauge("Lookups", float64(memStats.Lookups))
	collectGauge("MCacheInuse", float64(memStats.MCacheInuse))
	collectGauge("MCacheSys", float64(memStats.MCacheSys))
	collectGauge("MSpanInuse", float64(memStats.MSpanInuse))
	collectGauge("MSpanSys", float64(memStats.MSpanSys))
	collectGauge("Mallocs", float64(memStats.Mallocs))
	collectGauge("NextGC", float64(memStats.NextGC))
	collectGauge("NumForcedGC", float64(memStats.NumForcedGC))
	collectGauge("NumGC", float64(memStats.NumGC))
	collectGauge("OtherSys", float64(memStats.OtherSys))
	collectGauge("PauseTotalNs", float64(memStats.PauseTotalNs))
	collectGauge("StackInuse", float64(memStats.StackInuse))
	collectGauge("StackSys", float64(memStats.StackSys))
	collectGauge("Sys", float64(memStats.Sys))
	collectGauge("TotalAlloc", float64(memStats.TotalAlloc))

	logger.Log.Debug().Msg("done metrics collection routine")
}
