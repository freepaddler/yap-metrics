package config

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/caarlos0/env/v8"
	flag "github.com/spf13/pflag"

	"github.com/freepaddler/yap-metrics/internal/pkg/crypt"
	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
)

const (
	defaultPollInterval   = 2
	defaultReportInterval = 10
	defaultServerAddress  = "127.0.0.1:8080"
	defaultHTTPTimeout    = 5 * time.Second
	defaultLogLevel       = "info"
	defaultKey            = ""
	defaultRateLimit      = 1
)

// Config implements agent configuration
type Config struct {
	PollInterval    uint32        `env:"POLL_INTERVAL"`
	ReportInterval  uint32        `env:"REPORT_INTERVAL"`
	ServerAddress   string        `env:"ADDRESS"`
	HTTPTimeout     time.Duration `env:"HTTP_TIMEOUT"`
	LogLevel        string        `env:"LOG_LEVEL"`
	Key             string        `env:"KEY"`
	ReportRateLimit int           `env:"RATE_LIMIT"`
	PprofAddress    string        `env:"PPROF_ADDRESS"`
	PublicKeyFile   string        `env:"CRYPTO_KEY"`
	PublicKey       *rsa.PublicKey
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
		"d",
		defaultLogLevel,
		"logging `level` (trace, debug, info, warning, error)",
	)
	flag.StringVarP(
		&c.Key,
		"key",
		"k",
		defaultKey,
		"key for integrity hash calculation `secretkey`",
	)
	flag.IntVarP(
		&c.ReportRateLimit,
		"rateLimit",
		"l",
		defaultRateLimit,
		"max `number` of simultaneous reporters",
	)
	flag.StringVarP(
		&c.PprofAddress,
		"pprofAddress",
		"f",
		"",
		"enable an run pprof http server on `host:port`",
	)
	flag.StringVarP(
		&c.PublicKeyFile,
		"-crypto-key",
		"c",
		"",
		"`path` to public key file in PEM format",
	)

	flag.Parse()

	// env vars
	if err := env.Parse(&c); err != nil {
		logger.Log().Warn().Err(err).Msg("failed to parse ENV")
	}

	if c.ReportInterval < c.PollInterval {
		logger.Log().Fatal().Msgf("Report interval should be greater or equal to Poll interval")
	}

	if c.ReportRateLimit < 1 {
		logger.Log().Fatal().Msgf("Reporting rate limit should be greater than 0")
	}

	// check timeout
	if c.HTTPTimeout.Seconds() < 0.5 || c.HTTPTimeout.Seconds() > 999 {
		logger.Log().Warn().Msgf(
			"invalid httpTimeout value %s. Using default %s",
			c.HTTPTimeout.String(),
			defaultHTTPTimeout.String(),
		)
		c.HTTPTimeout = defaultHTTPTimeout
	}

	// print config
	logger.Log().Info().Interface("config", c).Msg("done config")

	// after print to avoid printing key in config
	// try to load public key
	if c.PublicKeyFile != "" {
		pFile, err := os.Open(c.PublicKeyFile)
		if err != nil {
			fmt.Printf("unable to open public key file: %s\n", err)
			os.Exit(1)
		}
		c.PublicKey, err = crypt.ReadPublicKey(pFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	return &c
}
