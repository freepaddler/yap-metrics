package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/freepaddler/yap-metrics/internal/store"
)

func main() {
	fmt.Println("Starting server...")

	// create new server instance with storage engine
	srv := &MetricsServer{storage: store.NewMemStorage()}

	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", srv.UpdateHandler)
	r.Get("/value/{type}/{name}", srv.ValueHandler)
	r.Get("/", srv.IndexHandler)

	//mux := http.NewServeMux()
	//mux.HandleFunc("/update/", srv.UpdateHandler)
	if err := http.ListenAndServe("localhost:8080", r); err != nil {
		panic(err)
	}

	// FIXME: this is never reachable until process control implementation
	fmt.Println("Stopping server...")
}
