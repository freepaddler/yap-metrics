package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

var (
	L zerolog.Logger
)

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	zerolog.DurationFieldUnit = time.Millisecond
	consoleLog := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		NoColor:    false,
		TimeFormat: time.DateTime + ".000",
	}
	L = zerolog.New(consoleLog).With().Timestamp().Caller().Logger()
}

func SetLevel(s string) {
	v, err := zerolog.ParseLevel(s)
	if err != nil {
		L.Warn().Err(err).Msg("invalid log level specified")
	}
	zerolog.SetGlobalLevel(v)
}
