package config

import (
	"errors"
	"io"
	"os"
	"time"

	"github.com/caarlos0/env/v8"
	flag "github.com/spf13/pflag"

	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
)

const (
	defaultPollInterval   = 2
	defaultReportInterval = 10
	defaultServerAddress  = "127.0.0.1:8080"
	defaultHTTPTimeout    = 5 * time.Second
	defaultLogLevel       = "info"
	defaultKey            = ""
)

// Config implements agent configuration
type Config struct {
	PollInterval   uint32        `env:"POLL_INTERVAL"`
	ReportInterval uint32        `env:"REPORT_INTERVAL"`
	ServerAddress  string        `env:"ADDRESS"`
	HTTPTimeout    time.Duration `env:"HTTP_TIMEOUT"`
	LogLevel       string        `env:"LOG_LEVEL"`
	Key            string        `env:"KEY""`
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
	flag.StringVarP(
		&c.ServerAddress,
		"serverAddress",
		"a",
		defaultServerAddress,
		"metrics collector server address `HOST:PORT`",
	)
	flag.Uint32VarP(
		&c.ReportInterval,
		"reportInterval",
		"r",
		defaultReportInterval,
		"how often to send metrics to server in `seconds`",
	)
	flag.Uint32VarP(
		&c.PollInterval,
		"pollInterval",
		"p",
		defaultPollInterval,
		"how often to collect metrics in `seconds`",
	)
	flag.DurationVarP(
		&c.HTTPTimeout,
		"httpTimeout",
		"t",
		defaultHTTPTimeout,
		"Metrics server response `timeout` min: 0.5s max: 999s",
	)
	flag.StringVarP(
		&c.LogLevel,
		"loglevel",
		"l",
		defaultLogLevel,
		"logging `level` (trace, debug, info, warning, error)",
	)
	flag.StringVarP(
		&c.Key,
		"key",
		"k",
		defaultServerAddress,
		"key for integrity hash calculation `secretkey`",
	)

	flag.Parse()

	// env vars
	if err := env.Parse(&c); err != nil {
		logger.Log.Warn().Err(err).Msg("failed to parse ENV")
	}

	if c.ReportInterval < c.PollInterval {
		logger.Log.Fatal().Msgf("Report interval should be greater or equal to Poll interval")
	}

	// check
	if c.HTTPTimeout.Seconds() < 0.5 || c.HTTPTimeout.Seconds() > 999 {
		logger.Log.Warn().Msgf(
			"invalid httpTimeout value %s. Using default %s",
			c.HTTPTimeout.String(),
			defaultHTTPTimeout.String(),
		)
		c.HTTPTimeout = defaultHTTPTimeout
	}

	return &c
}
