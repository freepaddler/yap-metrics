// Package controller contains agent business logic. It implements safe operations over MemoryStore and
// can properly restore flushed metrics values back in storage
package controller

import (
	"sync"
	"time"

	"github.com/freepaddler/yap-metrics/internal/pkg/models"
	"github.com/freepaddler/yap-metrics/internal/pkg/store"
)

type MetricsController struct {
	store      store.MemoryStore
	mu         sync.RWMutex
	countersTs map[string]time.Time // timestamps of counters updates
	gaugesTs   map[string]time.Time // timestamps of gauges updates
}

func New(storage store.MemoryStore) *MetricsController {
	return &MetricsController{
		store:    storage,
		gaugesTs: make(map[string]time.Time),
	}
}

func (mc *MetricsController) CollectCounter(name string, val int64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.store.IncCounter(name, val)
}

func (mc *MetricsController) CollectGauge(name string, val float64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.store.SetGauge(name, val)
	mc.gaugesTs[name] = time.Now()
}

// ReportAll gets all metrics from store and flushes it
func (mc *MetricsController) ReportAll() ([]models.Metrics, time.Time) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.store.Snapshot(true), time.Now()
}

// RestoreReport gets all metrics from store and flushes it
func (mc *MetricsController) RestoreReport(metrics []models.Metrics, ts time.Time) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	for _, v := range metrics {
		switch v.Type {
		case models.Gauge:
			// gauge value should always be latest
			if ts.After(mc.gaugesTs[v.Name]) {
				mc.store.SetGauge(v.Name, *v.FValue)
				mc.gaugesTs[v.Name] = ts
			}
		case models.Counter:
			// counter always increments
			mc.store.IncCounter(v.Name, *v.IValue)
		}
	}
}
