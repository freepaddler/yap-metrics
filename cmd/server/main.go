package main

import (
	"fmt"
	"net/http"

	"github.com/freepaddler/yap-metrics/internal/server/config"
	"github.com/freepaddler/yap-metrics/internal/server/handler"
	"github.com/freepaddler/yap-metrics/internal/server/router"
	"github.com/freepaddler/yap-metrics/internal/store"
)

func main() {
	// server configuration
	conf := config.NewConfig()
	fmt.Printf("Starting server at %s...\n", conf.Address)

	// let's define app composition
	//
	// server is:
	// storage - to operate data, should be an interface that implements all action on data
	// handlers - to access storage
	// router - to route requests to handlers
	//
	// dependencies:
	// storage()
	// handlers(storage)
	// router(handlers)

	// storage is interface, which methods should be called by handlers
	// router must call handlers

	// create new storage instance
	storage := store.NewMemStorage()
	// create http handlers instance
	httpHandlers := handler.NewHttpHandlers(storage)
	// create http router
	httpRouter := router.NewHttpRouter(httpHandlers)

	if err := http.ListenAndServe(conf.Address, httpRouter); err != nil {
		panic(err)
	}

	// FIXME: this is never reachable until process control implementation
	fmt.Println("Stopping server...")
}
