package router

import (
	"github.com/go-chi/chi/v5"

	"github.com/freepaddler/yap-metrics/internal/logger"
	"github.com/freepaddler/yap-metrics/internal/server/handler"
)

func NewHTTPRouter(h *handler.HTTPHandlers) *chi.Mux {
	r := chi.NewRouter()
	r.Use(logger.LogRequestResponse())
	r.Get("/", h.IndexMetricHandler)
	r.Route("/update", func(r chi.Router) {
		r.Post("/", h.UpdateMetricJSONHandler)
		r.Post("/{type}/{name}/{value}", h.UpdateMetricHandler)
	})
	r.Route("/value", func(r chi.Router) {
		r.Post("/", h.GetMetricJSONHandler)
		r.Get("/{type}/{name}", h.GetMetricHandler)
	})

	//r.Post("/update/{type}/{name}/{value}", h.UpdateMetricHandler)
	//r.Get("/value/{type}/{name}", h.GetMetricHandler)
	//r.Get("/", h.IndexMetricHandler)
	// TODO: we definitely need middleware with context to check request json
	//r.Post("/update", h.UpdateMetricJSONHandler)
	//r.Post("/value", h.GetMetricJSONHandler)

	return r
}
