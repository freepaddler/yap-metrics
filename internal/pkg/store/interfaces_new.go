// Package store describes metrics storage interfaces
package store

//go:generate mockgen -source $GOFILE -package=mocks -destination ../../../mocks/store_mock.go

import (
	"github.com/freepaddler/yap-metrics/internal/pkg/models"
)

// Gauge1 interface implements base operations on metrics with type gauge
type Gauge1 interface {
	// SetGauge sets Gauge value
	SetGauge(name string, value float64) float64
	// GetGauge returns Gauge value
	GetGauge(name string) (float64, bool)
	// DelGauge deletes Gauage
	DelGauge(name string)
}

// Counter1 interface implements base operations on metric with type counter
type Counter1 interface {
	// IncCounter increases Counter on passed value and returns increased one
	IncCounter(name string, value int64) int64
	// GetCounter returns Counter value
	GetCounter(name string) (int64, bool)
	// DelCounter deletes Counter
	DelCounter(name string)
}

// MemoryStore is a realtime metrics storage
// agent and server operate with this storage
type MemoryStore interface {
	Gauge1
	Counter1
	// Snapshot creates storage snapshot and returns it
	// if flush is true, stored metrics are deleted
	Snapshot(flush bool) []models.Metrics
}
