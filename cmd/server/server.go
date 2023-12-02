package main

import (
	"context"
	"crypto/rsa"
	"fmt"
	_ "net/http/pprof"
	"net/netip"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/freepaddler/yap-metrics/internal/app/server"
	"github.com/freepaddler/yap-metrics/internal/app/server/config"
	"github.com/freepaddler/yap-metrics/internal/app/server/handler"
	"github.com/freepaddler/yap-metrics/internal/app/server/router"
	"github.com/freepaddler/yap-metrics/internal/pkg/compress"
	"github.com/freepaddler/yap-metrics/internal/pkg/crypt"
	"github.com/freepaddler/yap-metrics/internal/pkg/ipmatcher"
	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
	"github.com/freepaddler/yap-metrics/internal/pkg/sign"
	"github.com/freepaddler/yap-metrics/internal/pkg/store"
	"github.com/freepaddler/yap-metrics/internal/pkg/store/filedump"
	"github.com/freepaddler/yap-metrics/internal/pkg/store/memory"
	"github.com/freepaddler/yap-metrics/internal/pkg/store/postgres"
)

var (
	// go build -ldflags " \
	// -X 'main.buildVersion=$(git describe --tag 2>/dev/null)' \
	// -X 'main.buildDate=$(date)' \
	// -X 'main.buildCommit=$(git rev-parse --short HEAD)' \
	// "
	buildVersion, buildDate, buildCommit = "N/A", "N/A", "N/A"
)

func main() {
	exitCode := 0
	defer func() { os.Exit(exitCode) }()

	fmt.Fprintf(
		os.Stdout,
		`Build version: %s
Build date: %s
Build commit %s
`, buildVersion, buildDate, buildCommit)

	// server configuration
	conf := config.NewConfig()

	// set log level
	logger.SetLevel(conf.LogLevel)

	// notify context
	nCtx, nStop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer nStop()

	var privateKey *rsa.PrivateKey
	if conf.PrivateKeyFile != "" {
		f, err := os.Open(conf.PrivateKeyFile)
		if err != nil {
			logger.Log().Error().Err(err).Msgf("unable to open private key file: %s", conf.PrivateKeyFile)
			exitCode = 2
			return
		}
		privateKey, err = crypt.ReadPrivateKey(f)
		if err != nil {
			logger.Log().Error().Err(err).Msgf("unable to read public key from file: %s", conf.PrivateKeyFile)
			exitCode = 2
			return
		}
	}

	// file dump
	var dump server.Dumper
	if conf.FileStoragePath != "" {
		f, err := os.OpenFile(conf.FileStoragePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
		if err != nil {
			logger.Log().Err(err).Msgf("can't open storage file %s", conf.FileStoragePath)
		} else {
			defer f.Close()
			dump = filedump.NewFileDump(f)
		}
	}

	// define storage
	var metricsStore store.Store
	if conf.DBURL != "" {
		var err error
		pgstore, err := postgres.NewPostgresStorage(conf.DBURL,
			postgres.WithTimeout(2*time.Second),
			postgres.WithRetry(1),
		)
		if err != nil {
			logger.Log().Err(err).Msg("unable to setup db storage")
			metricsStore = memory.NewMemoryStore()
		} else {
			defer pgstore.Close()
			metricsStore = pgstore
			// disable file dump if db is ok
			dump = nil
		}
	}
	if metricsStore == nil {
		metricsStore = memory.NewMemoryStore()
	}
	storage := store.NewStorageController(metricsStore)

	// define http handlers
	httpHandlers := handler.NewHTTPHandlers(storage)

	// parse trusted subnet
	var trustedSubnetEnable bool
	trustedSubnet := new(netip.Prefix)
	if conf.TrustedSubnet != "" {
		var err error
		if *trustedSubnet, err = netip.ParsePrefix(conf.TrustedSubnet); err != nil {
			logger.Log().Warn().Err(err).Msgf("unable to parse trusted subnet %s", conf.TrustedSubnet)
			*trustedSubnet = netip.MustParsePrefix("0.0.0.0/0")
		} else {
			trustedSubnetEnable = true
		}
	}
	//fmt.Println(trustedSubnetEnable)

	// setup router
	httpRouter := router.New(
		router.WithHandler(httpHandlers),
		router.WithLog(logger.LogRequestResponse),
		router.WithGunzip(compress.GunzipMiddleware),
		router.WithGzip(middleware.Compress(4, "application/json", "text/html")),
		router.WithCrypt(crypt.DecryptMiddleware(privateKey)),
		router.WithSign(sign.Middleware(conf.Key)),
		router.WithProfilerAt("/debug/"),
		router.WithIpMatcher(ipmatcher.IPMatchMiddleware(trustedSubnetEnable, *trustedSubnet)),
	)

	// init and run server
	app := server.NewServer(
		server.WithAddress(conf.Address),
		server.WithRouter(httpRouter),
		server.WithDump(dump),
		server.WithStorage(storage),
		server.WithDumpInterval(conf.StoreInterval),
		server.WithRestore(conf.Restore),
	)
	err := app.Run(nCtx)
	if err != nil {
		logger.Log().Error().Err(err).Msg("unclean exit")
		exitCode = 2
	}
	logger.Log().Info().Msg("server stopped")
}
