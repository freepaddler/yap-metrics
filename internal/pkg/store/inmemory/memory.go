// Package inmemory implements in-memory metrics storage primitives functions.
// it is NOT concurrent safe.
package inmemory

import (
	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
	"github.com/freepaddler/yap-metrics/internal/pkg/models"
	"github.com/freepaddler/yap-metrics/internal/pkg/store"
)

// Store is in-memory metric store structure
type Store struct {
	counters map[string]int64   // metrics of type counter
	gauges   map[string]float64 // metrics of type gauge
}

// New is a constructor for Store
func New() *Store {
	ms := &Store{}
	ms.counters = make(map[string]int64)
	ms.gauges = make(map[string]float64)
	return ms
}

// Gauge interface implementation
var _ store.Gauge1 = (*Store)(nil)

// SetGauge creates or updates gauge metric value in storage by its name
func (ms *Store) SetGauge(name string, fValue float64) float64 {
	ms.gauges[name] = fValue
	logger.Log().Debug().Msgf("SetGauge: store value %f for gauge %s", fValue, name)
	return fValue

}

// GetGauge returns gauge metric value and existence flag by its name
func (ms *Store) GetGauge(name string) (float64, bool) {
	v, ok := ms.gauges[name]
	return v, ok
}

// DelGauge removes gauge metric from storage by its name
func (ms *Store) DelGauge(name string) {
	delete(ms.gauges, name)
}

// Counter interface implementation
var _ store.Counter1 = (*Store)(nil)

// IncCounter creates new or increments counter metric value in storage by its name
func (ms *Store) IncCounter(name string, iValue int64) int64 {
	ms.counters[name] += iValue
	logger.Log().Debug().Msgf("IncCounter: add increment %d for counter %s", iValue, name)
	// pointer to map will not work
	v := ms.counters[name]
	return v
}

// GetCounter returns counter metric value and existence flag by its name
func (ms *Store) GetCounter(name string) (int64, bool) {
	v, ok := ms.counters[name]
	return v, ok
}

// DelCounter removes gauge metric from storage by its name
func (ms *Store) DelCounter(name string) {
	delete(ms.counters, name)
}

// MemoryStore interface implementation
var _ store.MemoryStore = (*Store)(nil)

// Snapshot returns all current memory store metrics in sorted by name slice
func (ms *Store) Snapshot(flush bool) []models.Metrics {
	// arrays of snapshot values (not pointers)
	counterV := make([]int64, len(ms.counters))
	gaugesV := make([]float64, len(ms.gauges))
	set := make([]models.Metrics, len(ms.counters)+len(ms.gauges))
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
	return set
}
