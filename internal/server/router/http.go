package router

import (
	"github.com/go-chi/chi/v5"

	"github.com/freepaddler/yap-metrics/internal/server/handler"
)

func NewHTTPRouter(h *handler.HTTPHandlers) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/update/{type}/{name}/{value}", h.UpdateMetricHandler)
	r.Get("/value/{type}/{name}", h.GetMetricHandler)
	r.Get("/", h.IndexMetricHandler)

	return r
}
