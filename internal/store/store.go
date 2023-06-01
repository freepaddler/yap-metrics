package store

// Methods for Gauge metrics
type Gauge interface {
	GaugeUpdate(name string, value float64)
}

// Methods for Counter metrics
type Counter interface {
	CounterUpdate(name string, value int64)
}

type Storage interface {
	Gauge
	Counter
}

// MemStorage stores metrics in memory store
type MemStorage struct {
	counters map[string]int64
	gauges   map[string]float64
}

// NewMemStorage is a constructor for MemStorage
func NewMemStorage() *MemStorage {
	ms := MemStorage{}
	ms.counters = make(map[string]int64)
	ms.gauges = make(map[string]float64)
	return &ms
}

func (ms *MemStorage) GaugeUpdate(name string, value float64) {
	ms.gauges[name] = value
}

func (ms *MemStorage) CounterUpdate(name string, value int64) {
	ms.counters[name] += value
}
