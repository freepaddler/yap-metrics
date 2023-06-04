package agent

import (
	"github.com/freepaddler/yap-metrics/internal/models"
)

// Counter interface for counters
type Counter interface {
	// Inc increase counter on value
	Inc(value int64)
	//// Get returns counter delta from last successful report
	//Get() int64
	//// SetReported sets last successful reported value
	//SetReported(value int64)
}

// Gauge interface for gauges
type Gauge interface {
	// Update set gauge value
	Update(value float64)
	//// Get returns gauge value and timestamp if updated from last successful report
	//Get() (value float64, ts time.Time)
	//// SetReported sets last successful reported timestamp
	//SetReported(t time.Time)
}

// Reporter interface implements sending metrics
type Reporter interface {
	// Report sends metric and returns true on success report
	Report(m models.Metrics) bool
}
