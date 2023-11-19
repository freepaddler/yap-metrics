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
	defaultAddress         = "127.0.0.1:8080"
	defaultLogLevel        = "info"
	defaultStoreInterval   = 300
	defaultFileStoragePath = "/tmp/metrics-db.json"
	defaultRestore         = true
	defaultDBURL           = ""
	defaultKey             = ""
)

// Config implements server configuration
type Config struct {
	Address         string `env:"ADDRESS"`
	LogLevel        string `env:"LOG_LEVEL"`
	StoreInterval   int    `env:"STORE_INTERVAL" json:"store_interval"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" json:"store_file"`
	Restore         bool   `env:"RESTORE"`
	UseFileStorage  bool   `json:"-"`
	DBURL           string `env:"DATABASE_DSN" json:"database_dsn"`
	UseDB           bool   `json:"-"`
	Key             string `env:"KEY"`
	PrivateKeyFile  string `env:"CRYPTO_KEY" json:"crypto_key"`
	ConfigFile      string `env:"CONFIG"`
}

// UnmarshalJSON to convert duration from config to uint32
func (c *Config) UnmarshalJSON(data []byte) error {
	type _conf Config
	_c := &struct {
		*_conf
		StoreInterval string `json:"store_interval"`
	}{
		_conf: (*_conf)(c),
	}
	if err := json.Unmarshal(data, _c); err != nil {
		return err
	}
	si, err := time.ParseDuration(_c.StoreInterval)
	if err != nil {
		return err
	}
	c.StoreInterval = int(si.Seconds())
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
	flag.StringVarP(&c.Address, "address", "a", defaultAddress, "server listening address `HOST:PORT`")
	flag.StringVarP(&c.LogLevel, "loglevel", "l", defaultLogLevel, "logging `level` (trace, debug, info, warning, error)")
	flag.IntVarP(&c.StoreInterval, "storeInterval", "i", defaultStoreInterval, "store to file interval in `seconds`")
	flag.StringVarP(&c.FileStoragePath, "fileStoragePath", "f", defaultFileStoragePath, "`path` to storage file")
	flag.BoolVarP(&c.Restore, "restore", "r", defaultRestore, "restore metrics after server start: `=true/false`")
	flag.StringVarP(&c.DBURL, "dbUri", "d", defaultDBURL, "database `uri` i.e. postgres://user:password@host:port/db")
	flag.StringVarP(&c.Key, "key", "k", defaultKey, "key for integrity hash calculation `secretkey`")
	flag.StringVarP(&c.PrivateKeyFile, "-crypto-key", "", "", "`path` to private key file in PEM format")
	flag.StringVarP(&c.ConfigFile, "config", "c", "", "`path` to configuration file in JSON format")
	flag.Parse()

	// env vars
	if err := env.Parse(&c); err != nil {
		logger.Log().Warn().Err(err).Msg("Failed to parse ENV")
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

	fsp, ok := os.LookupEnv("FILE_STORAGE_PATH")
	if ok {
		c.FileStoragePath = fsp
	}

	// print config
	printConfig(c)

	return &c
}
