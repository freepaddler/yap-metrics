// Package store describes metrics storage interfaces
package store

import (
	"context"

	"github.com/freepaddler/yap-metrics/internal/pkg/models"
)

// Gauge interface implements base operations on metrics with type gauge
type Gauge interface {
	// SetGauge sets Gauge value
	SetGauge(name string, value float64)
	// GetGauge returns Gauge value
	GetGauge(name string) (*float64, bool)
	// DelGauge deletes Gauage
	DelGauge(name string)
}

// Counter interface implements base operations on metric with type counter
type Counter interface {
	// IncCounter increases Counter on passed value and returns increased one
	IncCounter(name string, value int64) int64
	// GetCounter returns Counter value
	GetCounter(name string) (*int64, bool)
	// DelCounter deletes Counter
	DelCounter(name string)
}

// MemoryStore is a realtime metrics storage
// agent and server operate with this storage
type MemoryStore interface {
	Gauge
	Counter
	// Snapshot creates storage snapshot and returns it
	// if flush is true, collected metrics are deleted
	Snapshot(flush bool) []models.Metrics
}

// Storage represents realtime metrics storage
// Actually agent and server operate with this storage
type Storage interface {
	Gauge
	Counter
	// Snapshot creates storage snapshot and returns it
	// if flush is true, then metrics from snapshot are removed from Storage
	Snapshot(flush bool) []models.Metrics
	// GetMetric returns requested metric with actual value
	GetMetric(metric *models.Metrics) (bool, error)
	// UpdateMetrics updates metrics in store according to metric type, returns invalid metrics slice
	// if overwrite is true, then overwrite metric value instead of update
	UpdateMetrics(m []models.Metrics, overwrite bool)
	// RegisterHooks registers hooks which will be called on metrics update
	RegisterHooks(fns ...func([]models.Metrics))
}

// PersistentStorage represents persistent storage for metrics to be restored after server startup
type PersistentStorage interface {
	// RestoreStorage gets all latest metrics values from PersistentStorage and overwrites to Storage
	RestoreStorage(Storage)
	// SaveMetrics saves metrics to PersistentStorage
	// requires context to be able to be cancelled in main program loop
	SaveMetrics(context.Context, []models.Metrics)
	// SaveStorage saves all metrics from Storage to PersistentStorage
	SaveStorage(Storage)
	// Close stops and closes PersistentStorage
	Close()
	// Ping checks if PersistentStorage is accessible
	Ping() error
}
