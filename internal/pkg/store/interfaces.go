package store

//go:generate mockgen -source $GOFILE -package=mocks -destination ../../../mocks/Store_mock.go

import (
	"github.com/freepaddler/yap-metrics/internal/pkg/models"
)

// Gauge implements basic operations on metrics with type gauge
type Gauge interface {
	// SetGauge sets Gauge_old value
	SetGauge(name string, value float64) float64
	// GetGauge returns Gauge_old value
	GetGauge(name string) (float64, bool)
	// DelGauge deletes Gauage
	DelGauge(name string)
}

// Counter implements basic operations on metric with type counter
type Counter interface {
	// IncCounter increases Counter_old on passed value and returns increased one
	IncCounter(name string, value int64) int64
	// GetCounter returns Counter_old value
	GetCounter(name string) (int64, bool)
	// DelCounter deletes Counter_old
	DelCounter(name string)
}

// Store defines methods that should be implemented by any store.
type Store interface {
	Gauge
	Counter
	// Snapshot creates storage snapshot and returns it
	// if flush is true, stored metrics are deleted
	Snapshot(flush bool) []models.Metrics
	Ping() error
}
