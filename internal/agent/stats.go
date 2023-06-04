package agent

import (
	"time"

	"github.com/freepaddler/yap-metrics/internal/models"
)

// TODO: mutex/atomic for counters integrity

// Counter metric type
type counter struct {
	prev  int64
	value int64
}

func newCounter() *counter {
	return &counter{}
}

// Inc increases counter value
func (c *counter) Inc(v int64) {
	c.value += v
}

// get returns Metrics to report with increment = diff between value and reported
func (c *counter) get() (m models.Metrics) {
	m.Type = models.Counter
	m.Increment = c.value - c.prev
	m.Value = c.value
	return m
}

// Report sends counter Increment and updates last reported value
func (c *counter) Report(n string, r Reporter) {
	m := c.get()
	m.Name = n
	if r.Report(m) {
		c.reported(m.Value)
	}
	return
}

// reported updates prev with value after successful update
func (c *counter) reported(v int64) {
	if v > c.prev {
		c.prev = v
	}
}

// Gauge metric type
type gauge struct {
	gauge    float64
	updateTs time.Time
	reportTs time.Time
}

func newGauge() *gauge {
	return &gauge{}
}

// Update update gauge metric
func (g *gauge) Update(v float64) {
	g.gauge = v
	g.updateTs = time.Now()
}

// get returns Metrics to report
func (g *gauge) get() (m models.Metrics, ok bool) {
	if g.updateTs.After(g.reportTs) {
		m.Gauge = g.gauge
		m.Type = models.Gauge
		m.TimeStamp = g.updateTs
		ok = true
	}
	return
}

// Report sends gauge Value and updates last reported timestamp
func (g *gauge) Report(n string, r Reporter) {
	m, ok := g.get()
	if !ok {
		return
	}
	m.Name = n
	if r.Report(m) {
		g.reported(m.TimeStamp)
	}
	return
}

// reported updates reportedTs with updatedTs after successful report
func (g *gauge) reported(t time.Time) {
	if t.After(g.reportTs) {
		g.reportTs = t
	}
}
