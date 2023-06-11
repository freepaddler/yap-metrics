package config

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/caarlos0/env/v8"
	flag "github.com/spf13/pflag"
)

const (
	defaultAddress = "127.0.0.1:8080"
)

// Config implements server configuration
type Config struct {
	Address string `env:"ADDRESS"`
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
	flag.Parse()

	// env vars
	if err := env.Parse(&c); err != nil {
		fmt.Println("Error while parsing ENV", err)
	}

	return &c
}
