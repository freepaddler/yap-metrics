package models

const (
	Counter = "counter"
	Gauge   = "gauge"
)

type Metrics struct {
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	GaugeVal   float64 `json:"gauge,omitempty"`
	CounterVal int64   `json:"counter,omitempty"`
}
