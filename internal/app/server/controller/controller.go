package controller

import (
	"errors"
	"sync"
	"time"

	"github.com/freepaddler/yap-metrics/internal/pkg/models"
	"github.com/freepaddler/yap-metrics/internal/pkg/store"
)

var (
	ErrMetricNotFound = errors.New("metric not found in store")
)

// MetricsController implements server business logic layer
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

// GetAll returns all metrics from store
func (mc *MetricsController) GetAll() []models.Metrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.store.Snapshot(false)
}

// GetOne gets metric value from store and updates requested metric with it
func (mc *MetricsController) GetOne(metric *models.Metrics) error {
	switch metric.Type {
	case models.Counter:
		mc.mu.RLock()
		defer mc.mu.RUnlock()
		v, ok := mc.store.GetCounter(metric.Name)
		if !ok {
			return ErrMetricNotFound
		}
		metric.IValue = &v
	case models.Gauge:
		mc.mu.RLock()
		defer mc.mu.RUnlock()
		v, ok := mc.store.GetGauge(metric.Name)
		if !ok {
			return ErrMetricNotFound
		}
		metric.FValue = &v
	default:
		return models.ErrBadMetric
	}
	return nil
}

// updateMetric creates or updates metric in store, set new value to requested metric.
// It is a helper function to be called from public methods. It is not write safe.
func (mc *MetricsController) updateMetric(metric *models.Metrics) error {
	switch metric.Type {
	case models.Counter:
		v := mc.store.IncCounter(metric.Name, *metric.IValue)
		metric.IValue = &v
	case models.Gauge:
		v := mc.store.SetGauge(metric.Name, *metric.FValue)
		mc.gaugesTS[metric.Name] = time.Now()
		metric.FValue = &v
	default:
		return models.ErrBadMetric
	}
	return nil
}

// UpdateOne updates one metric in store, set new value to requested metric.
func (mc *MetricsController) UpdateOne(metric *models.Metrics) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	return mc.updateMetric(metric)
}

// UpdateMany updates batch of metric in store, set new value to requested metrics.
func (mc *MetricsController) UpdateMany(metrics []models.Metrics) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	for _, v := range metrics {
		if err := mc.updateMetric(&v); err != nil {
			return err
		}
	}
	return nil
}
