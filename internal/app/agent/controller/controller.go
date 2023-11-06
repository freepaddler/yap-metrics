// Package controller contains agent business logic. It implements safe operations over MemoryStore and
// can properly restore flushed metrics values back in storage
package controller

import (
	"sync"
	"time"

	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
	"github.com/freepaddler/yap-metrics/internal/pkg/models"
	"github.com/freepaddler/yap-metrics/internal/pkg/store"
)

// MetricsController implements agent business logic layer
type MetricsController struct {
	store    store.MemoryStore
	mu       sync.RWMutex
	gaugesTS map[string]time.Time // timestamps of gauges updates
}

// New is a MetricsController constructor
func New(storage store.MemoryStore) *MetricsController {
	return &MetricsController{
		store:    storage,
		gaugesTS: make(map[string]time.Time),
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
	mc.gaugesTS[name] = time.Now()
}

// ReportAll gets all metrics from store and flushes it
func (mc *MetricsController) ReportAll() ([]models.Metrics, time.Time) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.store.Snapshot(true), time.Now()
}

// RestoreReport gets all metrics from store and flushes it
func (mc *MetricsController) RestoreReport(metrics []models.Metrics, ts time.Time) {
	logger.Log().Debug().Msg("restoring report back to store")
	mc.mu.Lock()
	defer mc.mu.Unlock()
	for _, v := range metrics {
		switch v.Type {
		case models.Gauge:
			// gauge value should always be latest
			if ts.After(mc.gaugesTS[v.Name]) {
				mc.store.SetGauge(v.Name, *v.FValue)
				mc.gaugesTS[v.Name] = ts
			} else {
				logger.Log().Debug().Msgf("skip gauge '%s' restore, have newer value", v.Name)
			}
		case models.Counter:
			// counter always increments
			mc.store.IncCounter(v.Name, *v.IValue)
		}
	}
}
