package agent

// Counter interface for counters
type Counter interface {
	// Inc increase counter on value
	Inc(nValue int64)
}

// Gauge interface for gauges
type Gauge interface {
	// Update set gauge value
	Update(fValue float64)
}

// Reporter interface implements sending metrics
type Reporter interface {
	// Report sends metric and returns true on success report
	Report()
}
