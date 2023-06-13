package models

const (
	Counter = "counter"
	Gauge   = "gauge"
)

// Metrics model
type Metrics struct {
	Name   string  // metric name
	Type   string  // metric type (from constants)
	FValue float64 // float values for metrics
	IValue int64   // int values for metrics
}
