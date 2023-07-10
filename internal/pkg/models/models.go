package models

const (
	Counter = "counter"
	Gauge   = "gauge"
)

// Metrics model
type Metrics struct {
	Name   string   `json:"id"`
	Type   string   `json:"type"`
	FValue *float64 `json:"value,omitempty"`
	IValue *int64   `json:"delta,omitempty"`
}
