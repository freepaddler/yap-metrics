package agent

// StatsCollector is storage for collected metrics
type StatsCollector struct {
	gauges   map[string]*gauge
	counters map[string]*counter
}

func NewStatsCollector() *StatsCollector {
	return &StatsCollector{
		gauges:   make(map[string]*gauge),
		counters: make(map[string]*counter),
	}
}

func (sc *StatsCollector) Counter(name string) Counter {
	c, ok := sc.counters[name]
	if !ok {
		sc.counters[name] = newCounter()
		return sc.counters[name]
	}
	return c
}

func (sc *StatsCollector) Gauge(name string) Gauge {
	g, ok := sc.gauges[name]
	if !ok {
		sc.gauges[name] = newGauge()
		return sc.gauges[name]
	}
	return g
}

func (sc *StatsCollector) Report(r Reporter) {
	for k := range sc.counters {
		sc.counters[k].Report(k, r)
	}
	for k := range sc.gauges {
		sc.gauges[k].Report(k, r)
	}
}
