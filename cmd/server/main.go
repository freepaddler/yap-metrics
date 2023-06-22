package main

import (
	"net/http"

	"github.com/freepaddler/yap-metrics/internal/logger"
	"github.com/freepaddler/yap-metrics/internal/server/config"
	"github.com/freepaddler/yap-metrics/internal/server/handler"
	"github.com/freepaddler/yap-metrics/internal/server/router"
	"github.com/freepaddler/yap-metrics/internal/store/memory"
)

func main() {
	// global logger
	l := &logger.L
	// server configuration
	conf := config.NewConfig()
	// set log level

	l.Info().Msgf("Starting http server at %s...", conf.Address)

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
	storage := memory.NewMemStorage()
	// create http handlers instance
	httpHandlers := handler.NewHTTPHandlers(storage)
	// create http router
	httpRouter := router.NewHTTPRouter(httpHandlers)

	if err := http.ListenAndServe(conf.Address, httpRouter); err != nil {
		l.Fatal().Err(err).Msg("unable to start http server")
	}

	// FIXME: this is never reachable until process control implementation
	l.Info().Msg("Stopping server...")
}
