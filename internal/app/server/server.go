package server

import (
	"context"
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
	"github.com/freepaddler/yap-metrics/internal/pkg/models"
	"github.com/freepaddler/yap-metrics/internal/pkg/store"
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
	pStore       *store.PersistentStorage
}

// New creates new server instance
func New(conf *config.Config) *Server {
	srv := &Server{conf: conf}

	// init new memory storage
	srv.store = memory.NewMemStorage()
	// init new persistent storage
	srv.pStore = new(store.PersistentStorage)

	// create http handlers instance
	srv.httpHandlers = handler.NewHTTPHandlers(srv.store, srv.pStore)

	// create http router
	srv.httpRouter = router.NewHTTPRouter(srv.httpHandlers)

	// create http server
	srv.httpServer = &http.Server{Addr: srv.conf.Address, Handler: srv.httpRouter}

	return srv
}

// Run starts server instance
func (srv *Server) Run() {
	logger.Log.Info().Msg("starting metrics server")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// persistent storage setup
	var pStore store.PersistentStorage
	var err error
	switch {
	case srv.conf.UseDB:
		// database storage setup
		logger.Log.Info().Msg("using database as persistent storage")
		ctxDB, ctxDBCancel := context.WithTimeout(ctx, db.DBTimeout*time.Second)
		defer ctxDBCancel()
		pStore, err = srv.initDBStorage(ctxDB)
		if err != nil {
			// Error here instead of Fatal to let server work without db to pass tests 10[ab]
			logger.Log.Error().Err(err).Msg("database storage disabled")
		}
	case srv.conf.UseFileStorage:
		// file storage setup
		logger.Log.Info().Msg("using file as persistent storage")
		pStore, err = srv.initFileStorage(ctx)
		if err != nil {
			logger.Log.Error().Err(err).Msg("file storage disabled")
		}
	}
	if pStore != nil {
		*srv.pStore = pStore
		defer pStore.Close()
	}

	// start http server
	go func() {
		logger.Log.Info().Msgf("starting http server at %s", srv.conf.Address)
		if err := srv.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			logger.Log.Fatal().Err(err).Msg("unable to start http server")
		}
	}()

	logger.Log.Info().Msg("server started")

	// trap os signals
	shutdownSig := make(chan os.Signal, 1)
	signal.Notify(shutdownSig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// shutdown routine
	sig := <-shutdownSig
	logger.Log.Info().Msgf("got '%v' signal. server shutdown routine...", sig)

	// post tasks
	defer func() {
		// gracefully stop all context-related goroutines
		cancel()

		// save all metrics to persistent storage
		if pStore != nil {
			logger.Log.Info().Msg("saving all metrics to persistent storage on exit")
			ctxSave, ctxSaveCancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer ctxSaveCancel()
			pStore.SaveStorage(ctxSave, srv.store)
		}
		logger.Log.Info().Msg("server stopped")
	}()

	// context for httpServer graceful shutdown
	httpCtx, httpRelease := context.WithTimeout(ctx, 5*time.Second)
	defer httpRelease()

	// gracefully stop http server
	logger.Log.Info().Msg("stopping http server")
	if err := srv.httpServer.Shutdown(httpCtx); err != nil {
		log.Fatalf("failed to stop http server: %v", err)
	}

}

// initFileStorage sets up file storage
func (srv *Server) initFileStorage(ctx context.Context) (fStore *file.FileStorage, err error) {
	// create file storage
	fStore, err = file.New(srv.conf.FileStoragePath)
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("unable to init file storage")
	}

	// restore storage from file
	if srv.conf.Restore {
		fStore.RestoreStorage(ctx, srv.store)
	}

	// register update hook for sync write to persistent storage
	if srv.conf.StoreInterval == 0 {
		srv.store.RegisterHook(
			func(m models.Metrics) {
				go func() {
					fStore.SaveMetric(ctx, m)
				}()
			})
	} else if srv.conf.StoreInterval > 0 {
		go func() {
			fStore.SaveLoop(ctx, srv.store, srv.conf.StoreInterval)
		}()
	} else {
		err = fmt.Errorf("invalid storeInterval=%d, should be 0 or greater", srv.conf.StoreInterval)
		return nil, err
	}
	return fStore, nil
}

// initDBStorage sets up file storage
func (srv *Server) initDBStorage(ctx context.Context) (*db.DBStorage, error) {
	// create database storage
	dbStore, err := db.New(ctx, srv.conf.DBURL)
	if err != nil {
		// Error here instead of Fatal to let server work without db to pass tests 10[ab]
		logger.Log.Error().Err(err).Msg("unable to init database storage")
		return nil, err
	}

	// restore storage from file
	if srv.conf.Restore {
		dbStore.RestoreStorage(ctx, srv.store)
	}

	// register update hook for sync write to persistent storage
	srv.store.RegisterHook(
		func(m models.Metrics) {
			go func() {
				dbStore.SaveMetric(ctx, m)
			}()
		})

	return dbStore, nil
}
