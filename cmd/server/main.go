package main

import (
	"fmt"
	"net/http"

	"github.com/freepaddler/yap-metrics/internal/store"
)

func main() {
	fmt.Println("Starting server...")

	// create new server instance with storage engine
	srv := &MetricsServer{storage: store.NewMemStorage()}

	mux := http.NewServeMux()
	mux.HandleFunc("/update/", srv.UpdateHandler)
	if err := http.ListenAndServe("localhost:8080", mux); err != nil {
		panic(err)
	}

	// FIXME this is never reachable until process control implementation
	fmt.Println("Stopping server...")
}
