package main

import (
	"net/http"
	"time"

	"github.com/freepaddler/yap-metrics/internal/logger"
	"github.com/freepaddler/yap-metrics/internal/server/config"
	"github.com/freepaddler/yap-metrics/internal/server/handler"
	"github.com/freepaddler/yap-metrics/internal/server/router"
	"github.com/freepaddler/yap-metrics/internal/store/file"
	"github.com/freepaddler/yap-metrics/internal/store/memory"
)

func main() {
	// server configuration
	conf := config.NewConfig()
	// set log level
	logger.SetLevel(conf.LogLevel)
	logger.Log.Debug().Interface("Config", conf).Msg("done config")
	logger.Log.Info().Msgf("Starting http server at %s...", conf.Address)

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

	// create file storage
	fStore, err := file.NewFileStorage(conf.FileStoragePath, conf.StoreInterval)
	if err != nil {
		logger.Log.Error().Err(err).Msg("unable to init file storage")
	}
	defer fStore.Close()
	// create new storage instance
	storage := memory.NewMemStorage()
	// restore storage from file
	if fStore != nil && conf.Restore {
		fStore.RestoreStorage(storage)
	}
	// register update hook for sync write to persistent storage
	if fStore != nil && conf.StoreInterval == 0 {
		storage.RegisterHook(fStore.Updated)
	}
	// create http handlers instance
	httpHandlers := handler.NewHTTPHandlers(storage)
	// create http router
	httpRouter := router.NewHTTPRouter(httpHandlers)

	go func() {
		if err := http.ListenAndServe(conf.Address, httpRouter); err != nil {
			logger.Log.Fatal().Err(err).Msg("unable to start http server")
		}
	}()

	logger.Log.Debug().Msg("Starting file storage loop...")
	ticker := 1
	for {
		if fStore != nil &&
			conf.StoreInterval > 0 &&
			ticker%conf.StoreInterval == 0 {
			fStore.SaveStorage(storage)
		}
		time.Sleep(time.Second)
		ticker++
	}

	// FIXME: this is never reachable until process control implementation
	// logger.Log.Info().Msg("Stopping server...")
}
