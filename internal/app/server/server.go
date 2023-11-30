package server

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
	"github.com/freepaddler/yap-metrics/internal/pkg/models"
)

type Dumper interface {
	Dump(metrics []models.Metrics) error
	Restore() (metrics []models.Metrics, lastDump time.Time, err error)
}

type DumpStorage interface {
	GetAll() []models.Metrics
	RestoreLatest(metrics []models.Metrics, ts time.Time)
}

// Server represents server application
type Server struct {
	httpServer   *http.Server
	wg           sync.WaitGroup
	dump         Dumper
	dumpInterval time.Duration
	storage      DumpStorage
	restore      bool
}

func NewServer(opts ...func(server *Server)) *Server {
	srv := &Server{
		httpServer: &http.Server{},
	}
	for _, o := range opts {
		o(srv)
	}
	return srv
}

func WithAddress(a string) func(server *Server) {
	return func(server *Server) {
		server.httpServer.Addr = a
	}
}

func WithRouter(r http.Handler) func(server *Server) {
	return func(server *Server) {
		server.httpServer.Handler = r
	}
}

func WithStorage(s DumpStorage) func(server *Server) {
	return func(server *Server) {
		server.storage = s
	}
}

func WithDump(d Dumper) func(server *Server) {
	return func(server *Server) {
		server.dump = d
	}
}

func WithDumpInterval(interval int) func(server *Server) {
	return func(server *Server) {
		server.dumpInterval = time.Duration(interval) * time.Second
	}
}

func WithRestore(r bool) func(server *Server) {
	return func(server *Server) {
		server.restore = r
	}
}

// New creates new server instance
//func New(conf *config.Config) *Server {
//	srv := &Server{conf: conf}
//
//	// TODO: remove
//	// init new memory storage
//	srv.store1 = memory.NewMemStorage()
//
//	// FIXED
//
//	srv.storage = inmemory.New()
//	srv.controller = store.NewStorageController(srv.storage)
//
//	// create http handlers instance
//	srv.httpHandlers = handler.NewHTTPHandlers(srv.controller)
//
//	// create http router
//	//srv.httpRouter = router.NewHTTPRouter(srv.httpHandlers, conf.Key, conf.PrivateKey)
//
//	// create http server
//	srv.httpServer = &http.Server{Addr: srv.conf.Address, Handler: srv.httpRouter}
//
//	return srv
//}

// checkCtxCancel validates context is not cancelled, exit fatal if it is
func checkCtxCancel(ctx context.Context) {
	if err := ctx.Err(); err != nil {
		logger.Log().Fatal().Err(err).Msg("application terminated")
	}
}

// Run starts server instance
func (srv *Server) Run(ctx context.Context) error {
	checkCtxCancel(ctx)
	logger.Log().Info().Msg("starting metrics server")

	// file storage setup
	if srv.dump != nil {
		if srv.restore {
			logger.Log().Info().Msg("restoring metrics from file")
			// restore may last long, we don't need to wait
			go func() {
				metrics, ts, err := srv.dump.Restore()
				if err != nil {
					logger.Log().Err(err).Msg("unable to restore from dump")
				} else {
					logger.Log().Info().Msgf("found dump from %s, restoring", ts.Format(time.UnixDate))
					srv.storage.RestoreLatest(metrics, ts)
				}
			}()
		}
		if srv.dumpInterval > 0 {
			logger.Log().Info().Msgf("start metrics dump every %.f seconds", srv.dumpInterval.Seconds())
			srv.wg.Add(1)
			go func() {
				defer srv.wg.Done()
				for {
					select {
					case <-time.After(srv.dumpInterval):
						logger.Log().Debug().Msg("run dump")
						metrics := srv.storage.GetAll()
						srv.dump.Dump(metrics)
					case <-ctx.Done():
						logger.Log().Info().Msg("stop metrics dump")
						return
					}
				}
			}()
		}
	}

	// start http server
	srv.wg.Add(1)
	go func() {
		defer srv.wg.Done()
		logger.Log().Info().Msgf("starting http server at %s", srv.httpServer.Addr)
		if err := srv.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Log().Fatal().Err(err).Msg("unable to start http server")
		}
		logger.Log().Info().Msg("http server stopped acquiring new connections")
	}()

	time.Sleep(500 * time.Millisecond)
	logger.Log().Info().Msg("server started")

	// waiting for main context to be cancelled
	<-ctx.Done()
	logger.Log().Info().Msg("got stop request. stopping server")

	// context for httpServer graceful shutdown
	httpCtx, httpRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer httpRelease()

	// gracefully stop http server
	logger.Log().Info().Msg("stopping http server")
	if err := srv.httpServer.Shutdown(httpCtx); err != nil {
		logger.Log().Err(err).Msg("failed to stop http server gracefully. force stop")
		_ = srv.httpServer.Close()
	}
	logger.Log().Info().Msg("http server stopped")

	// wait until tasks stopped
	srv.wg.Wait()

	// shutdown tasks

	// save all metrics to persistent storage
	if srv.dump != nil {
		logger.Log().Info().Msg("dump metrics to file on exit")
		metrics := srv.storage.GetAll()
		srv.dump.Dump(metrics)
	}

	return nil
}
