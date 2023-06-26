package router

import (
	"github.com/go-chi/chi/v5"

	"github.com/freepaddler/yap-metrics/internal/logger"
	"github.com/freepaddler/yap-metrics/internal/server/handler"
)

func NewHTTPRouter(h *handler.HTTPHandlers) *chi.Mux {
	r := chi.NewRouter()
	r.Use(logger.LogRequestResponse())
	r.Post("/update/{type}/{name}/{value}", h.UpdateMetricHandler)
	//	r.Post("/update/", h.BunchUpdateMetricHandler)
	r.Get("/value/{type}/{name}", h.GetMetricHandler)
	r.Get("/", h.IndexMetricHandler)

	return r
}
