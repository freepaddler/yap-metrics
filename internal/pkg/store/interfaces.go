package store

import (
	"context"

	"github.com/freepaddler/yap-metrics/internal/pkg/models"
)

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
	Flush()
	GetMetric(metric *models.Metrics) (bool, error)
	SetMetric(metric *models.Metrics) error
	RegisterHook(fns ...func([]models.Metrics))
}

type PersistentStorage interface {
	// RestoreStorage gets all latest metrics from PersistentStorage and writes to Storage
	RestoreStorage(context.Context, Storage)
	// SaveMetric saves metrics to PersistentStorage
	SaveMetric(context.Context, []models.Metrics)
	// SaveStorage saves all metrics from Storage to PersistentStorage
	SaveStorage(context.Context, Storage)
	// Close stops and closes PersistentStorage
	Close()
	// Ping checks if PersistentStorage is accessible
	Ping() error
}
