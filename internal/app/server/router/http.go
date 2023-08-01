package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/freepaddler/yap-metrics/internal/app/server/handler"
	"github.com/freepaddler/yap-metrics/internal/pkg/compress"
	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
	"github.com/freepaddler/yap-metrics/internal/pkg/sign"
)

func NewHTTPRouter(h *handler.HTTPHandlers, key string) *chi.Mux {
	r := chi.NewRouter()
	r.Use(logger.LogRequestResponse)
	r.Use(compress.GunzipMiddleware)
	r.Use(middleware.Compress(4, "application/json", "text/html"))
	r.Use(sign.Middleware(key))
	r.Get("/", h.IndexMetricHandler)
	r.Route("/update", func(r chi.Router) {
		r.Post("/", h.UpdateMetricJSONHandler)
		r.Post("/{type}/{name}/{value}", h.UpdateMetricHandler)
	})
	r.Route("/value", func(r chi.Router) {
		r.Post("/", h.GetMetricJSONHandler)
		r.Get("/{type}/{name}", h.GetMetricHandler)
	})
	r.Get("/ping", h.PingHandler)
	r.Route("/updates", func(r chi.Router) {
		r.Post("/", h.UpdateMetricsBatchHandler)
	})

	return r
}
