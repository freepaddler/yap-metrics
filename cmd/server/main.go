package main

import (
	"net/http"

	"github.com/freepaddler/yap-metrics/internal/store"
)

func main() {
	// create new server instance with storage engine
	srv := &MetricsServer{storage: store.NewMemStorage()}

	mux := http.NewServeMux()
	mux.Handle("/update/", ValidateMetricURL(http.HandlerFunc(srv.UpdateHandler)))
	if err := http.ListenAndServe("localhost:8080", mux); err != nil {
		panic(err)
	}
}

// MetricsServer instance
type MetricsServer struct {
	storage store.Storage
}
