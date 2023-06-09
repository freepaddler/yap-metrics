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
	FValue    float64 // float values for metrics
	Increment int64
	IValue    int64 // int values for metrics
	TimeStamp time.Time
}
