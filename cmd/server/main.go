package main

import (
	_ "net/http/pprof"

	"github.com/freepaddler/yap-metrics/internal/app/server"
	"github.com/freepaddler/yap-metrics/internal/app/server/config"
	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
)

func main() {
	// server configuration
	conf := config.NewConfig()

	// set log level
	logger.SetLevel(conf.LogLevel)

	// print running config
	logger.Log.Info().Interface("config", conf).Msg("done config")

	// init and run server
	app := server.New(conf)
	app.Run()
}
