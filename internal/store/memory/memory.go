package memory

import (
	"errors"
	"sort"

	"github.com/freepaddler/yap-metrics/internal/logger"
	"github.com/freepaddler/yap-metrics/internal/models"
)

// MemStorage is in-memory metric store
type MemStorage struct {
	counters map[string]int64
	gauges   map[string]float64
}

func (ms *MemStorage) GetAllMetrics() []models.Metrics {
	// make values arrays
	counterV := make([]int64, len(ms.counters))
	gaugesV := make([]float64, len(ms.gauges))
	set := make([]models.Metrics, 0)
	for name, value := range ms.counters {
		counterV = append(counterV, value)
		set = append(set, models.Metrics{Type: models.Counter, Name: name, IValue: &counterV[len(counterV)-1]})
	}
	for name, value := range ms.gauges {
		gaugesV = append(gaugesV, value)
		set = append(set, models.Metrics{Type: models.Gauge, Name: name, FValue: &gaugesV[len(gaugesV)-1]})
	}
	sort.Slice(set, func(i, j int) bool {
		return set[i].Name < set[j].Name
	})
	return set
}

// NewMemStorage is a constructor for MemStorage
func NewMemStorage() *MemStorage {
	ms := &MemStorage{}
	ms.counters = make(map[string]int64)
	ms.gauges = make(map[string]float64)
	return ms
}

func (ms *MemStorage) SetGauge(name string, iValue float64) {
	ms.gauges[name] = iValue
	logger.Log.Debug().Msgf("SetGauge: store value %f for gauge %s", iValue, name)
}

func (ms *MemStorage) GetGauge(name string) (*float64, bool) {
	v, ok := ms.gauges[name]
	return &v, ok
}

func (ms *MemStorage) DelGauge(name string) {
	delete(ms.gauges, name)
}

func (ms *MemStorage) IncCounter(name string, iValue int64) {
	ms.counters[name] += iValue
	logger.Log.Debug().Msgf("IncCounter: add increment %d for counter %s", iValue, name)
}

func (ms *MemStorage) GetCounter(name string) (*int64, bool) {
	v, ok := ms.counters[name]
	return &v, ok
}

func (ms *MemStorage) DelCounter(name string) {
	delete(ms.counters, name)
}

// GetMetric searches requested metric in store
// and updates requested metric with value from store
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

// SetMetric updates metric value in store
// then, updates passed metric value with new store value
func (ms *MemStorage) SetMetric(m *models.Metrics) error {
	var err error
	switch m.Type {
	case models.Gauge:
		ms.SetGauge(m.Name, *m.FValue)
	case models.Counter:
		ms.IncCounter(m.Name, *m.IValue)
	default:
		err = errors.New("invalid metric type")
		return err
	}
	found, err := ms.GetMetric(m)
	if err == nil && !found {
		err = errors.New("just updated metric not found")
	}
	return err
}
