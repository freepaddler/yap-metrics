package main

import (
	"fmt"
	"net/http"

	"github.com/caarlos0/env/v8"
	"github.com/go-chi/chi/v5"
	flag "github.com/spf13/pflag"

	"github.com/freepaddler/yap-metrics/internal/store"
)

const (
	defaultAddress = "127.0.0.1:8080"
)

// global configuration
type config struct {
	Address string `env:"ADDRESS"`
}

func main() {
	// global config
	var conf config

	// cmd params
	flag.StringVarP(&conf.Address, "address", "a", defaultAddress, "server listening address HOST:PORT")
	flag.Parse()

	// env vars
	if err := env.Parse(&conf); err != nil {
		fmt.Println("Error while parsing ENV", err)
	}

	fmt.Printf("Starting server at %s...\n", conf.Address)

	// create new server instance with storage engine
	srv := &MetricsServer{storage: store.NewMemStorage()}

	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", srv.UpdateMetricHandler)
	r.Get("/value/{type}/{name}", srv.GetMetricHandler)
	r.Get("/", srv.IndexMetricHandler)

	if err := http.ListenAndServe(conf.Address, r); err != nil {
		panic(err)
	}

	// FIXME: this is never reachable until process control implementation
	fmt.Println("Stopping server...")
}
