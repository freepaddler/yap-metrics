package logger

import (
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

var Log *zerolog.Logger

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	zerolog.DurationFieldUnit = time.Millisecond
	consoleLog := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		NoColor:    false,
		TimeFormat: time.DateTime + ".000",
	}
	l := zerolog.New(consoleLog).With().Timestamp().Caller().Logger()
	Log = &l
}

func SetLevel(s string) {
	v, err := zerolog.ParseLevel(s)
	if err != nil {
		Log.Warn().Err(err).Msg("invalid log level specified")
	}
	zerolog.SetGlobalLevel(v)
}

func LogRequestResponse(next http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {

		// to get response data
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		tStart := time.Now()
		defer func() {
			dur := time.Since(tStart)
			Log.Info().
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
