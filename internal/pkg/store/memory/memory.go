// Package memory implements in-memory metrics storage
package memory

// TODO: write examples

import (
	"errors"
	"sort"
	"sync"

	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
	"github.com/freepaddler/yap-metrics/internal/pkg/models"
)

// MemStorage is in-memory metric store structure
type MemStorage struct {
	mu       sync.RWMutex
	counters map[string]int64         // metrics of type counter
	gauges   map[string]float64       // metrics of type gauge
	hooks    []func([]models.Metrics) // notification hooks
}

// NewMemStorage is a constructor for MemStorage
func NewMemStorage() *MemStorage {
	ms := &MemStorage{}
	ms.counters = make(map[string]int64)
	ms.gauges = make(map[string]float64)
	return ms
}

// Gauge interface implementation
// var _ store.Gauge = (*MemStorage)(nil)

// SetGauge creates or updates gauge metric value in storage by its name
func (ms *MemStorage) SetGauge(name string, fValue float64) {
	ms.gauges[name] = fValue
	logger.Log().Debug().Msgf("SetGauge: store value %f for gauge %s", fValue, name)

}

// GetGauge returns gauge metric value and existence flag by its name
func (ms *MemStorage) GetGauge(name string) (*float64, bool) {
	v, ok := ms.gauges[name]
	return &v, ok
}

// DelGauge removes gauge metric from storage by its name
func (ms *MemStorage) DelGauge(name string) {
	delete(ms.gauges, name)
}

// Counter interface implementation
// var _ store.Counter = (*MemStorage)(nil)

// IncCounter creates new or increments counter metric value in storage by its name
func (ms *MemStorage) IncCounter(name string, iValue int64) int64 {
	ms.counters[name] += iValue
	logger.Log().Debug().Msgf("IncCounter: add increment %d for counter %s", iValue, name)
	// pointer to map will not work
	v := ms.counters[name]
	return v
}

// GetCounter returns counter metric value and existence flag by its name
func (ms *MemStorage) GetCounter(name string) (*int64, bool) {
	v, ok := ms.counters[name]
	return &v, ok
}

// DelCounter removes gauge metric from storage by its name
func (ms *MemStorage) DelCounter(name string) {
	delete(ms.counters, name)
}

// Storage interface implementation
//var _ store.Storage = (*MemStorage)(nil)

// Snapshot returns all current memory store metrics in sorted by name slice
func (ms *MemStorage) Snapshot(flush bool) []models.Metrics {
	// make values arrays
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	counterV := make([]int64, len(ms.counters))
	gaugesV := make([]float64, len(ms.gauges))
	set := make([]models.Metrics, 0)
	for name, value := range ms.counters {
		counterV = append(counterV, value)
		set = append(set, models.Metrics{Type: models.Counter, Name: name, IValue: &counterV[len(counterV)-1]})
		if flush {
			delete(ms.counters, name)
		}
	}
	for name, value := range ms.gauges {
		gaugesV = append(gaugesV, value)
		set = append(set, models.Metrics{Type: models.Gauge, Name: name, FValue: &gaugesV[len(gaugesV)-1]})
		if flush {
			delete(ms.gauges, name)
		}
	}
	sort.Slice(set, func(i, j int) bool {
		return set[i].Name < set[j].Name
	})
	return set
}

// GetMetric enriches requested in metric with it current value.
// Returns true if metric is found, false otherwise.
// Returns error if metric type is invalid
func (ms *MemStorage) GetMetric(m *models.Metrics) (bool, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	var found bool
	var err error
	switch m.Type {
	case models.Gauge:
		m.FValue, found = ms.GetGauge(m.Name)
	case models.Counter:
		m.IValue, found = ms.GetCounter(m.Name)
	default:
		err = errors.New("invalid metric type")
	}
	return found, err
}

// UpdateMetrics creates or updates batch of metrics.
// If overwrite is true, then counter metrics are not incremented, but set to the requested values.
func (ms *MemStorage) UpdateMetrics(m []models.Metrics, overwrite bool) {
	ms.mu.Lock()
	for i := 0; i < len(m); i++ {
		switch m[i].Type {
		case models.Gauge:
			ms.SetGauge(m[i].Name, *m[i].FValue)
		case models.Counter:
			if overwrite {
				ms.DelCounter(m[i].Name)
				ms.IncCounter(m[i].Name, *m[i].IValue)
			} else {
				*m[i].IValue = ms.IncCounter(m[i].Name, *m[i].IValue)
			}
		default:
			logger.Log().Warn().Msgf("UpdateMetrics: invalid metric '%s' type '%s', skipping", m[i].Name, m[i].Type)
		}
	}
	// call update persistent storage hooks
	ms.mu.Unlock()
	ms.updateHook(m)
}

// RegisterHooks registers functions to be called on metrics update
func (ms *MemStorage) RegisterHooks(fns ...func([]models.Metrics)) {
	ms.hooks = append(ms.hooks, fns...)
}

// updateHook notifies persistent storage that metric was updated
func (ms *MemStorage) updateHook(m []models.Metrics) {
	for _, hook := range ms.hooks {
		hook(m)
	}
}
