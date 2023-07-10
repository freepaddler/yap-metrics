package main

import (
	"github.com/freepaddler/yap-metrics/internal/app/agent"
	"github.com/freepaddler/yap-metrics/internal/app/agent/config"
	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
)

func main() {

	// agent configuration
	conf := config.NewConfig()

	// set log level
	logger.SetLevel(conf.LogLevel)

	// print running config
	logger.Log.Info().Interface("config", conf).Msg("done config")

	// init and run agent
	app := agent.New(conf)
	app.Run()

}
