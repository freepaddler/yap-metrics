package store

import (
	"fmt"

	"github.com/freepaddler/yap-metrics/internal/models"
)

// MemStorage is in-memory metric store
type MemStorage struct {
	counters map[string]int64
	gauges   map[string]float64
}

func (ms *MemStorage) GetAllMetrics() []models.Metrics {
	set := make([]models.Metrics, 0)
	for name, value := range ms.counters {
		set = append(set, models.Metrics{Type: models.Counter, Name: name, Value: value})
	}
	for name, value := range ms.gauges {
		set = append(set, models.Metrics{Type: models.Gauge, Name: name, Gauge: value})
	}
	return set
}

// NewMemStorage is a constructor for MemStorage
func NewMemStorage() *MemStorage {
	ms := &MemStorage{}
	ms.counters = make(map[string]int64)
	ms.gauges = make(map[string]float64)
	return ms
}

func (ms *MemStorage) SetGauge(name string, value float64) {
	ms.gauges[name] = value
	fmt.Printf("SetGauge: store value %f for gauge %s\n", value, name)
}

func (ms *MemStorage) GetGauge(name string) (float64, bool) {
	v, ok := ms.gauges[name]
	return v, ok
}

func (ms *MemStorage) IncCounter(name string, value int64) {
	ms.counters[name] += value
	fmt.Printf("IncCounter: add increment %d for counter %s\n", value, name)
}

func (ms *MemStorage) GetCounter(name string) (int64, bool) {
	v, ok := ms.counters[name]
	return v, ok
}
