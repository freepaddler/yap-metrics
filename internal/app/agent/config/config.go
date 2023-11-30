package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/caarlos0/env/v8"
	flag "github.com/spf13/pflag"

	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
)

const (
	defaultServerAddress  = "127.0.0.1:8080"
	defaultReportInterval = 10
	defaultPollInterval   = 2

	defaultHTTPTimeout = 5 * time.Second
	defaultLogLevel    = "info"
	defaultKey         = ""
	defaultRateLimit   = 1
)

// Config implements agent configuration
type Config struct {
	ServerAddress  string `env:"ADDRESS" json:"address"`
	ReportInterval uint32 `env:"REPORT_INTERVAL"`
	PollInterval   uint32 `env:"POLL_INTERVAL"`
	PublicKeyFile  string `env:"CRYPTO_KEY" json:"crypto_key"`

	HTTPTimeout     time.Duration `env:"HTTP_TIMEOUT"`
	LogLevel        string        `env:"LOG_LEVEL"`
	Key             string        `env:"KEY"`
	ReportRateLimit int           `env:"RATE_LIMIT"`
	PprofAddress    string        `env:"PPROF_ADDRESS"`

	ConfigFile string `env:"CONFIG"`
}

// UnmarshalJSON to convert duration from config to uint32
func (c *Config) UnmarshalJSON(data []byte) error {
	type _conf Config
	_c := &struct {
		*_conf
		PollInterval   string `json:"poll_interval"`
		ReportInterval string `json:"report_interval"`
	}{
		_conf: (*_conf)(c),
	}
	if err := json.Unmarshal(data, _c); err != nil {
		return err
	}
	pi, err := time.ParseDuration(_c.PollInterval)
	if err != nil {
		return err
	}
	c.PollInterval = uint32(pi.Seconds())
	ri, err := time.ParseDuration(_c.ReportInterval)
	if err != nil {
		return err
	}
	c.ReportInterval = uint32(ri.Seconds())
	return nil
}

func parseConfigFile(c *Config) error {
	if c.ConfigFile != "" {
		f, err := os.Open(c.ConfigFile)
		if err != nil {
			return err
		}
		data, err := io.ReadAll(f)
		if err != nil {
			return err
		}
		err = json.Unmarshal(data, c)
		if err != nil {
			return err
		}
	}
	return nil
}

func printConfig(c Config) {
	fmt.Println("Startup configuration:")
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		logger.Log().Error().Err(err).Msg("unable to parse config")
		return
	}
	fmt.Printf("%s\n", data)
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
		"",
		"",
		"`path` to public key file in PEM format",
	)
	flag.StringVarP(
		&c.ConfigFile,
		"config",
		"c",
		"",
		"`path` to configuration file in JSON format",
	)

	// parse Flags
	flag.Parse()

	// parse env vars
	if err := env.Parse(&c); err != nil {
		logger.Log().Warn().Err(err).Msg("failed to parse ENV")
	}

	// get settings from configuration file
	if c.ConfigFile != "" {
		if err := parseConfigFile(&c); err == nil {
			// re-read flags and env vars
			flag.Parse()
			env.Parse(&c)
		} else {
			logger.Log().Warn().Err(err).Msg("failed to parse config file")
		}
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
	printConfig(c)

	return &c
}
