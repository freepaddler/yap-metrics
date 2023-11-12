package config

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/caarlos0/env/v8"
	flag "github.com/spf13/pflag"

	"github.com/freepaddler/yap-metrics/internal/pkg/crypt"
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
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	UseFileStorage  bool
	DBURL           string `env:"DATABASE_DSN"`
	UseDB           bool
	Key             string `env:"KEY"`
	PrivateKeyFile  string `env:"CRYPTO_KEY"`
	PrivateKey      *rsa.PrivateKey
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
	flag.BoolVarP(&c.Restore, "restore", "r", defaultRestore, "restore metrcis after server start: `true/false`")
	flag.StringVarP(&c.DBURL, "dbUri", "d", defaultDBURL, "database `uri` i.e. postgres://user:password@host:port/db")
	flag.StringVarP(&c.Key, "key", "k", defaultKey, "key for integrity hash calculation `secretkey`")
	flag.StringVarP(
		&c.PrivateKeyFile,
		"-crypto-key",
		"c",
		"",
		"`path` to private key file in PEM format",
	)
	flag.Parse()

	// env vars
	if err := env.Parse(&c); err != nil {
		logger.Log().Warn().Err(err).Msg("Failed to parse ENV")
	}

	fsp, ok := os.LookupEnv("FILE_STORAGE_PATH")
	if ok {
		c.FileStoragePath = fsp
	}

	// choose persistent storage
	switch {
	case c.DBURL != "":
		c.UseDB = true
	case c.FileStoragePath != "":
		c.UseFileStorage = true
	}

	// print running config
	logger.Log().Info().Interface("config", c).Msg("done config")

	// after config printing to avoid private key in logs
	// try to load private key
	if c.PrivateKeyFile != "" {
		pFile, err := os.Open(c.PrivateKeyFile)
		if err != nil {
			fmt.Printf("unable to open private key file: %s\n", err)
			os.Exit(1)
		}
		c.PrivateKey, err = crypt.ReadPrivateKey(pFile)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	return &c
}
