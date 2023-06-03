package models

const (
	Counter = "counter"
	Gauge   = "gauge"
)

type Metrics struct {
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	Value     float64 `json:"value,omitempty"`
	Increment int64   `json:"increment,omitempty"`
}
