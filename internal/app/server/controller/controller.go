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
func (mc *MetricsController) GetOne(request models.MetricRequest) (m models.Metrics, err error) {
	switch request.Type {
	case models.Counter:
		mc.mu.RLock()
		defer mc.mu.RUnlock()
		v, ok := mc.store.GetCounter(request.Name)
		if !ok {
			err = ErrMetricNotFound
			return
		}
		m.IValue = &v
	case models.Gauge:
		mc.mu.RLock()
		defer mc.mu.RUnlock()
		v, ok := mc.store.GetGauge(request.Name)
		if !ok {
			err = ErrMetricNotFound
			return
		}
		m.FValue = &v
	default:
		err = models.ErrInvalidMetric
	}
	if err == nil {
		m.Name = request.Name
		m.Type = request.Type
	}
	return
}

// updateMetric creates or updates metric in store, set new value to requested metric.
// It is a helper function to be called from public methods. It is not write safe.
func (mc *MetricsController) updateMetric(metric *models.Metrics) error {
	switch metric.Type {
	case models.Counter:
		if metric.IValue == nil {
			return models.ErrInvalidMetric
		}
		v := mc.store.IncCounter(metric.Name, *metric.IValue)
		metric.IValue = &v
	case models.Gauge:
		if metric.FValue == nil {
			return models.ErrInvalidMetric
		}
		v := mc.store.SetGauge(metric.Name, *metric.FValue)
		mc.gaugesTS[metric.Name] = time.Now()
		metric.FValue = &v
	default:
		return models.ErrInvalidMetric
	}
	return nil
}

// TODO: question why not to use func with variadic args update(metrics ...*models.Metrics)?

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
