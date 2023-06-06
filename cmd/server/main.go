package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	flag "github.com/spf13/pflag"

	"github.com/freepaddler/yap-metrics/internal/store"
)

const (
	defaultAddress = "127.0.0.1:8080"
)

type config struct {
	address string
}

func main() {
	conf := config{}
	flag.StringVarP(&conf.address, "address", "a", defaultAddress, "server listening address HOST:PORT")
	flag.Parse()

	fmt.Printf("Starting server at %s...\n", conf.address)

	// create new server instance with storage engine
	srv := &MetricsServer{storage: store.NewMemStorage()}

	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", srv.UpdateHandler)
	r.Get("/value/{type}/{name}", srv.ValueHandler)
	r.Get("/", srv.IndexHandler)

	//mux := http.NewServeMux()
	//mux.HandleFunc("/update/", srv.UpdateHandler)
	if err := http.ListenAndServe(conf.address, r); err != nil {
		panic(err)
	}

	// FIXME: this is never reachable until process control implementation
	fmt.Println("Stopping server...")
}
