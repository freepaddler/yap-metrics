package config

import (
	"fmt"

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
	c := Config{}

	// cmd params
	flag.StringVarP(&c.Address, "address", "a", defaultAddress, "server listening address HOST:PORT")
	flag.Parse()

	// env vars
	if err := env.Parse(&c); err != nil {
		fmt.Println("Error while parsing ENV", err)
	}

	return &c
}
