package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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
	logger.Log.Debug().Interface("config", conf).Msg("done config")
	logger.Log.Info().Msgf("starting http server at %s...", conf.Address)

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

	httpServer := &http.Server{Addr: conf.Address, Handler: httpRouter}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal().Err(err).Msg("unable to start http server")
		}
	}()

	// shutdown gracefully
	gracefulShutdown := make(chan os.Signal, 1)
	signal.Notify(gracefulShutdown, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	sig := <-gracefulShutdown

	logger.Log.Info().Msgf("got '%v' signal. server stopping routine...", sig)
	logger.Log.Info().Msg("stopping http server")
	// TODO: replace with shutdown after context topic
	if err := httpServer.Close(); err != nil {
		logger.Log.Warn().Err(err).Msg("failed to stop http server")
	}
	if fStore != nil {
		logger.Log.Info().Msg("stopping file storage")
		fStore.SaveStorage(storage)
		fStore.Close()
	}
	logger.Log.Info().Msg("server stopped")

	// FIXME: this is never reachable until process control implementation
	// logger.Log.Info().Msg("stopping server...")
}

func initFileStorage(conf *config.Config, storage store.Storage) (fStore *file.FileStorage, err error) {
	if conf.UseFileStorage {
		// create file storage
		fStore, err = file.NewFileStorage(conf.FileStoragePath)
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
