package router

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"

	"github.com/freepaddler/yap-metrics/internal/logger"
	"github.com/freepaddler/yap-metrics/internal/server/handler"
)

var (
	l = &logger.L
)

func NewHTTPRouter(h *handler.HTTPHandlers) *chi.Mux {
	r := chi.NewRouter()
	r.Use(LogRequestResponse(l))
	r.Post("/update/{type}/{name}/{value}", h.UpdateMetricHandler)
	r.Get("/value/{type}/{name}", h.GetMetricHandler)
	r.Get("/", h.IndexMetricHandler)

	return r
}

func LogRequestResponse(logger *zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		logFn := func(w http.ResponseWriter, r *http.Request) {

			// to get response data
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			tStart := time.Now()
			defer func() {
				dur := time.Since(tStart)
				logger.Info().
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
}
