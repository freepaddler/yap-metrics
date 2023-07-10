package memory

import (
	"errors"
	"sort"
	"sync"

	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
	"github.com/freepaddler/yap-metrics/internal/pkg/models"
)

// MemStorage is in-memory metric store
type MemStorage struct {
	mu       sync.Mutex
	counters map[string]int64
	gauges   map[string]float64
	hooks    []func([]models.Metrics)
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

func (ms *MemStorage) SetGauge(name string, fValue float64) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.gauges[name] = fValue
	logger.Log.Debug().Msgf("SetGauge: store value %f for gauge %s", fValue, name)
	// pointer to map will not work
	v := fValue
	ms.updateHook([]models.Metrics{
		{
			Name:   name,
			Type:   models.Gauge,
			FValue: &v,
		}})
}
func (ms *MemStorage) GetGauge(name string) (*float64, bool) {
	v, ok := ms.gauges[name]
	return &v, ok
}
func (ms *MemStorage) DelGauge(name string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	delete(ms.gauges, name)
}

// Counter interface implementation
// var _ store.Counter = (*MemStorage)(nil)

func (ms *MemStorage) IncCounter(name string, iValue int64) int64 {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.counters[name] += iValue
	logger.Log.Debug().Msgf("IncCounter: add increment %d for counter %s", iValue, name)
	// pointer to map will not work
	v := ms.counters[name]
	return v
}
func (ms *MemStorage) GetCounter(name string) (*int64, bool) {
	v, ok := ms.counters[name]
	return &v, ok
}
func (ms *MemStorage) DelCounter(name string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	delete(ms.counters, name)
}

// Storage interface implementation
//var _ store.Storage = (*MemStorage)(nil)

func (ms *MemStorage) Snapshot(flush bool) []models.Metrics {
	// make values arrays
	ms.mu.Lock()
	defer ms.mu.Unlock()
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
func (ms *MemStorage) GetMetric(m *models.Metrics) (bool, error) {
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
func (ms *MemStorage) UpdateMetrics(m []models.Metrics, overwrite bool) []models.Metrics {
	invalid := make([]models.Metrics, 0)
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
			logger.Log.Warn().Msgf("UpdateMetrics: invalid metric '%s' type '%s', skipping", m[i].Name, m[i].Type)
			invalid = append(invalid, m[i])
			// remove invalid element from slice
			m[i] = m[len(m)-1]
			m = m[:len(m)-1]
			// return index back
			i--
		}
	}
	// call update persistent storage hooks
	ms.updateHook(m)

	return invalid
}
func (ms *MemStorage) RegisterHooks(fns ...func([]models.Metrics)) {
	ms.hooks = append(ms.hooks, fns...)
}

// updateHook notifies persistent storage that metric was updated
func (ms *MemStorage) updateHook(m []models.Metrics) {
	for _, hook := range ms.hooks {
		hook(m)
	}
}
