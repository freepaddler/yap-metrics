package config

import (
	"errors"
	"io"
	"os"

	"github.com/caarlos0/env/v8"
	flag "github.com/spf13/pflag"

	"github.com/freepaddler/yap-metrics/internal/logger"
)

const (
	defaultAddress  = "127.0.0.1:8080"
	defaultLogLevel = "info"
)

// Config implements server configuration
type Config struct {
	Address  string `env:"ADDRESS"`
	LogLevel string `env:"LOG_LEVEL"`
}

func NewConfig() *Config {
	var c Config
	var output io.Writer = os.Stderr
	flag.CommandLine.SetOutput(output)
	// sorting is based on long args, doesn't look too good
	flag.CommandLine.SortFlags = false
	// avoid message "pflag: help requested"
	flag.ErrHelp = errors.New("")

	// cmd params
	flag.StringVarP(&c.Address, "address", "a", defaultAddress, "server listening address `HOST:PORT`")
	flag.StringVarP(&c.LogLevel, "loglevel", "l", defaultLogLevel, "logging `level` (trace, debug, info, warning, error)")
	flag.Parse()

	// env vars
	if err := env.Parse(&c); err != nil {
		logger.Log.Warn().Err(err).Msg("failed to parse ENV")
	}

	logger.Log.Debug().Interface("Config", c).Msg("done config")

	return &c
}
