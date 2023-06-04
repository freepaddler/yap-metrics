package main

import "github.com/freepaddler/yap-metrics/internal/agent"

func reportMetrics(sc *agent.StatsCollector, r agent.Reporter) {
	sc.Report(r)
}
