package server

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/freepaddler/yap-metrics/internal/app/server/config"
	"github.com/freepaddler/yap-metrics/internal/app/server/handler"
	"github.com/freepaddler/yap-metrics/internal/app/server/router"
	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
	"github.com/freepaddler/yap-metrics/internal/pkg/store/db"
	"github.com/freepaddler/yap-metrics/internal/pkg/store/file"
	"github.com/freepaddler/yap-metrics/internal/pkg/store/memory"
)

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

// Server represents server application
type Server struct {
	conf         *config.Config
	store        *memory.MemStorage
	httpHandlers *handler.HTTPHandlers
	httpRouter   *chi.Mux
	httpServer   *http.Server
	db           *sql.DB
}

// New creates new server instance
func New(conf *config.Config) *Server {
	srv := &Server{conf: conf}

	// init new memory storage
	srv.store = memory.NewMemStorage()

	// create database connection
	if conf.UseDB {
		srv.db = db.New(conf.DBURL)
	}

	// create http handlers instance
	srv.httpHandlers = handler.NewHTTPHandlers(srv.store, srv.db)

	// create http router
	srv.httpRouter = router.NewHTTPRouter(srv.httpHandlers)

	// create http server
	srv.httpServer = &http.Server{Addr: srv.conf.Address, Handler: srv.httpRouter}

	return srv
}

// Run starts server instance
func (srv *Server) Run() {
	logger.Log.Info().Msg("starting metrics server")

	// trap os signals
	shutdownSig := make(chan os.Signal, 1)
	signal.Notify(shutdownSig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// file storage setup
	fStore, err := srv.initFileStorage(ctx)
	if err != nil {
		logger.Log.Error().Err(err).Msg("file storage disabled")
	} else {
		defer fStore.Close()
	}

	// start http server
	go func() {
		logger.Log.Info().Msgf("starting http server at %s", srv.conf.Address)
		if err := srv.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			logger.Log.Fatal().Err(err).Msg("unable to start http server")
		}
	}()

	logger.Log.Info().Msg("server started")

	// shutdown routine
	sig := <-shutdownSig
	logger.Log.Info().Msgf("got '%v' signal. server shutdown routine...", sig)

	// context for httpServer graceful shutdown
	httpCtx, httpRelease := context.WithTimeout(ctx, 5*time.Second)
	defer httpRelease()

	// post tasks
	defer func() {
		// gracefully stop all context-related goroutines
		cancel()

		// save all metrics to file storage
		if fStore != nil {
			logger.Log.Info().Msg("stopping file storage")
			fStore.SaveStorage(srv.store)
			fStore.Close()
		}
		logger.Log.Info().Msg("server stopped")
	}()

	// gracefully stop http server
	logger.Log.Info().Msg("stopping http server")
	if err := srv.httpServer.Shutdown(httpCtx); err != nil {
		log.Fatalf("failed to stop http server: %v", err)
	}

}

// initFileStorage sets up file storage
func (srv *Server) initFileStorage(ctx context.Context) (fStore *file.FileStorage, err error) {
	if srv.conf.UseFileStorage {
		// create file storage
		fStore, err = file.NewFileStorage(srv.conf.FileStoragePath)
		if err != nil {
			logger.Log.Fatal().Err(err).Msg("unable to init file storage")
		}

		// restore storage from file
		if srv.conf.Restore {
			fStore.RestoreStorage(srv.store)
		}

		// register update hook for sync write to persistent storage
		if srv.conf.StoreInterval == 0 {
			srv.store.RegisterHook(fStore.SaveMetric)
		} else if srv.conf.StoreInterval > 0 {
			go func() {
				fStore.SaveLoop(ctx, srv.store, srv.conf.StoreInterval)
			}()
		} else {
			err = fmt.Errorf("invalid storeInterval=%d, should be 0 or greater", srv.conf.StoreInterval)
			return nil, err
		}
	}
	return fStore, nil
}
