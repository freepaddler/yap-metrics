package stats

// TODO: mutex/atomic for counters integrity

// Counter metric type
type counter struct {
	value int64
}

// Inc increases counter value
func (c *counter) Inc(v int64) {
	c.value += v
}

// Gauge metric type
type gauge struct {
	gauge float64
}

// Update update gauge metric
func (g *gauge) Update(v float64) {
	g.gauge = v
}
