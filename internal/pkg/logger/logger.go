package logger

import (
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

const (
	logLevel zerolog.Level = zerolog.InfoLevel // default log level
)

var (
	log *zerolog.Logger
)

func init() {
	zerolog.SetGlobalLevel(logLevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	zerolog.DurationFieldUnit = time.Millisecond
	consoleLog := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		NoColor:    false,
		TimeFormat: time.DateTime + ".000",
	}
	l := zerolog.New(consoleLog).With().Timestamp().Caller().Logger()
	log = &l
}

func Log() *zerolog.Logger {
	return log
}

func SetLevel(s string) {
	v, err := zerolog.ParseLevel(s)
	if err != nil {
		log.Warn().Err(err).Msgf("invalid log level specified, using default level '%s'", logLevel)
	}
	log.Info().Msgf("set log level to '%s'", v)
	zerolog.SetGlobalLevel(v)
}

func LogRequestResponse(next http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {

		// to get response data
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		tStart := time.Now()
		defer func() {
			dur := time.Since(tStart)
			log.Info().
				Str("host", r.Host).
				Str("url", r.URL.Path).
				Str("method", r.Method).
				Dur("ms_served", dur).
				Int("status", ww.Status()).
				Int("bytes_sent", ww.BytesWritten()).
				Msg("http request")
		}()
		next.ServeHTTP(ww, r)
	}
	return http.HandlerFunc(logFn)
}
