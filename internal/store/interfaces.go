package store

import "github.com/freepaddler/yap-metrics/internal/models"

type Gauge interface {
	SetGauge(name string, value float64)
	GetGauge(name string) (*float64, bool)
	DelGauge(name string)
}

type Counter interface {
	IncCounter(name string, value int64)
	GetCounter(name string) (*int64, bool)
	DelCounter(name string)
}

type Storage interface {
	Gauge
	Counter
	Snapshot() []models.Metrics
	GetMetric(metric *models.Metrics) (bool, error)
	SetMetric(metric *models.Metrics) error
}

type PersistentStorage interface {
	RestoreStorage(storage Storage)
	SaveMetric(storage Storage, metric models.Metrics)
	SaveStorage(storage Storage)
}
