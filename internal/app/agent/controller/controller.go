// Package controller contains agent business logic. It implements safe operations over MemoryStore and
// can properly restore flushed metrics values back in storage
package controller

import (
	"fmt"
	"sync"
	"time"

	"github.com/freepaddler/yap-metrics/internal/pkg/models"
	"github.com/freepaddler/yap-metrics/internal/pkg/store"
)

var (
	ErrNotFound    = fmt.Errorf("metric is not found")
	ErrInvalidType = fmt.Errorf("invalid metric type")
)

type MetricsController struct {
	store      store.MemoryStore
	mu         sync.RWMutex
	countersTs map[string]time.Time // timestamps of counters updates
	gaugesTs   map[string]time.Time // timestamps of gauges updates
}

func New(storage store.MemoryStore) *MetricsController {
	return &MetricsController{
		store:      storage,
		countersTs: make(map[string]time.Time),
		gaugesTs:   make(map[string]time.Time),
	}
}

func (mc *MetricsController) CollectCounter(name string, val int64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.store.IncCounter(name, val)
	mc.countersTs[name] = time.Now()
}

func (mc *MetricsController) CollectGauge(name string, val float64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.store.SetGauge(name, val)
	mc.gaugesTs[name] = time.Now()
}

// getMetric updates requested metric with value from store
func (mc *MetricsController) getMetric(metric *models.Metrics) error {
	switch metric.Type {
	case models.Gauge:
		if val, ok := mc.store.GetGauge(metric.Name); ok {
			metric.FValue = &val
			return nil
		}
		return ErrNotFound
	case models.Counter:
		if val, ok := mc.store.GetCounter(metric.Name); ok {
			metric.IValue = &val
			return nil
		}
		return ErrNotFound
	default:
		return ErrInvalidType
	}
}

func (mc *MetricsController) storeMetric(metric *models.Metrics, ts time.Time) error {
	switch metric.Type {
	case models.Gauge:
		// if gauge was already updated there is no need to update it with old value
		if ts.After(mc.gaugesTs[metric.Name]) {
			val := mc.store.SetGauge(metric.Name, *metric.FValue)
			metric.FValue = &val
			mc.gaugesTs[metric.Name] = time.Now()
			return nil
		}
		return mc.getMetric(metric)
	case models.Counter:
		// counter always increments
		val := mc.store.IncCounter(metric.Name, *metric.IValue)
		metric.IValue = &val
		mc.countersTs[metric.Name] = time.Now()
		return nil
	default:
		return ErrInvalidType
	}
}

func (mc *MetricsController) storeMetrics(metrics []models.Metrics, ts time.Time) error {
	var errCount int
	for _, v := range metrics {
		err := mc.storeMetric(&v, ts)
		if err != nil {
			errCount++
		}
	}
	if errCount > 0 {
		return fmt.Errorf("Failed to store %d metrics", errCount)
	}
	return nil
}

func (mc *MetricsController) delMetric(metric models.Metrics) {
	switch metric.Type {
	case models.Gauge:
		mc.store.DelGauge(metric.Name)
	case models.Counter:
		mc.store.DelCounter(metric.Name)
	}
}

// StoreOne creates or updates one metric
func (mc *MetricsController) StoreOne(metric *models.Metrics) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	return mc.storeMetric(metric, time.Now())
}

// StoreMany creates or updates batch of metric
func (mc *MetricsController) StoreMany(metrics []models.Metrics) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	return mc.storeMetrics(metrics, time.Now())
}

// GetOne returns requested metric value
func (mc *MetricsController) GetOne(metric *models.Metrics) error {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.getMetric(metric)
}

// GetAll returns all metrics from store
func (mc *MetricsController) GetAll() []models.Metrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.store.Snapshot(false)
}

// ExtractOne gets one metric from store and deletes it
func (mc *MetricsController) ExtractOne(metric *models.Metrics) (time.Time, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	var ts time.Time
	err := mc.getMetric(metric)
	if err != nil {
		return ts, err
	}
	mc.delMetric(*metric)
	return time.Now(), nil
}

// RestoreOne restores metric
func (mc *MetricsController) RestoreOne(metric models.Metrics, ts time.Time) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.storeMetric(&metric, ts)
}

// ExtractAll gets all metrics from store and flushes it
func (mc *MetricsController) ExtractAll() ([]models.Metrics, time.Time) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.store.Snapshot(true), time.Now()
}

// RestoreAll restores batch of metrics
func (mc *MetricsController) RestoreAll(metrics []models.Metrics, ts time.Time) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	for _, v := range metrics {
		mc.storeMetric(&v, ts)
	}
}
