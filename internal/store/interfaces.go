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

	// TODO: question
	// нужно ли было на самом деле реализовывать методы GetMetric и SetMetric
	// вполне можно было обойтись методами самих метрик
	// зато Get/Set Metric позволяют реализовать все в одном месте

}

type PersistentStorage interface {
	RestoreStorage(storage Storage)
	Updated(storage Storage, metric models.Metrics)
	SaveStorage(storage Storage)
}
