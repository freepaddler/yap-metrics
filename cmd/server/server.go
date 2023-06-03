package main

import "github.com/freepaddler/yap-metrics/internal/store"

// MetricsServer instance
type MetricsServer struct {
	storage store.Storage
}
