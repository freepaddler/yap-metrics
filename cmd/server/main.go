package main

import (
	"fmt"
	"net/http"

	"github.com/freepaddler/yap-metrics/internal/logger"
	"github.com/freepaddler/yap-metrics/internal/server/config"
	"github.com/freepaddler/yap-metrics/internal/server/handler"
	"github.com/freepaddler/yap-metrics/internal/server/router"
	"github.com/freepaddler/yap-metrics/internal/store"
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

	// create new storage instance
	storage := memory.NewMemStorage()

	// TODO: question
	// вопрос к композиции package main
	// вижу 2 логичных варианта:
	// вариант 1: всю логику инициализации выносить в отдельные функции, в func main -
	// вызов этих функций
	// вариант 2: логику инициализации описывать в func main, и в результате вызывать func run(),
	// где, собственно, и должен происходить запуск сервисов сервера

	// file storage setup
	fStore, err := initFileStorage(conf, storage)
	if err != nil {
		logger.Log.Error().Err(err).Msg("file storage disabled")
	} else {
		defer fStore.Close()
	}

	// create http handlers instance
	httpHandlers := handler.NewHTTPHandlers(storage)
	// create http router
	httpRouter := router.NewHTTPRouter(httpHandlers)

	if err := http.ListenAndServe(conf.Address, httpRouter); err != nil {
		logger.Log.Fatal().Err(err).Msg("unable to start http server")
	}

	// FIXME: this is never reachable until process control implementation
	// logger.Log.Info().Msg("Stopping server...")
}

func initFileStorage(conf *config.Config, storage store.Storage) (*file.FileStorage, error) {
	var fStore *file.FileStorage
	if conf.UseFileStorage {
		// create file storage
		fStore, err := file.NewFileStorage(conf.FileStoragePath)
		if err != nil {
			logger.Log.Fatal().Err(err).Msg("unable to init file storage")
		}

		// restore storage from file
		if conf.Restore {
			fStore.RestoreStorage(storage)
		}

		// register update hook for sync write to persistent storage
		if conf.StoreInterval == 0 {
			storage.RegisterHook(fStore.SaveMetric)
		} else if conf.StoreInterval > 0 {
			go func() {
				fStore.SaveLoop(storage, conf.StoreInterval)
			}()
		} else {
			err = fmt.Errorf("invalid storeInterval=%d, should be 0 or greater", conf.StoreInterval)
			return nil, err
		}
	}
	return fStore, nil
}
