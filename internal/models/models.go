package models

import "time"

const (
	Counter = "counter"
	Gauge   = "gauge"
)

// Metrics model
type Metrics struct {
	Name      string
	Type      string
	Gauge     float64
	Increment int64
	Value     int64
	TimeStamp time.Time
}
