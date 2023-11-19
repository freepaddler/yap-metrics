package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type MetricsHTTPHandler interface {
	IndexMetricHandler(w http.ResponseWriter, _ *http.Request)
	GetMetricHandler(w http.ResponseWriter, r *http.Request)
	GetMetricJSONHandler(w http.ResponseWriter, r *http.Request)
	UpdateMetricHandler(w http.ResponseWriter, r *http.Request)
	UpdateMetricJSONHandler(w http.ResponseWriter, r *http.Request)
	UpdateMetricsBatchHandler(w http.ResponseWriter, r *http.Request)
	PingHandler(w http.ResponseWriter, r *http.Request)
}

type Middleware func(http.Handler) http.Handler

type Router struct {
	handler      MetricsHTTPHandler
	gzip         Middleware
	gunzip       Middleware
	log          Middleware
	crypt        Middleware
	sign         Middleware
	profilerPath string
}

func WithHandler(h MetricsHTTPHandler) func(router *Router) {
	return func(router *Router) {
		router.handler = h
	}
}

func WithGzip(mw Middleware) func(router *Router) {
	return func(router *Router) {
		router.gzip = mw
	}
}

func WithGunzip(mw Middleware) func(router *Router) {
	return func(router *Router) {
		router.gunzip = mw
	}
}

func WithLog(mw Middleware) func(router *Router) {
	return func(router *Router) {
		router.log = mw
	}
}

func WithSign(mw Middleware) func(router *Router) {
	return func(router *Router) {
		router.sign = mw
	}
}

func WithCrypt(mw Middleware) func(router *Router) {
	return func(router *Router) {
		router.crypt = mw
	}
}

func WithProfilerAt(path string) func(router *Router) {
	return func(router *Router) {
		router.profilerPath = path
	}
}

func New(opts ...func(router *Router)) http.Handler {
	router := new(Router)
	for _, o := range opts {
		o(router)
	}
	return router.create()
}

func (router Router) create() http.Handler {
	r := chi.NewRouter()
	// adds middleware if not nil
	mw := func(m Middleware) {
		if m != nil {
			r.Use(m)
		}
	}
	// middleware order is important
	mw(router.log)
	mw(router.gunzip)
	mw(router.gzip)
	mw(router.crypt)
	mw(router.sign)

	if router.profilerPath != "" {
		r.Mount(router.profilerPath, middleware.Profiler())
	}

	r.Get("/", router.handler.IndexMetricHandler)
	r.Route("/update", func(r chi.Router) {
		r.Post("/", router.handler.UpdateMetricJSONHandler)
		r.Post("/{type}/{name}/{value}", router.handler.UpdateMetricHandler)
	})
	r.Route("/value", func(r chi.Router) {
		r.Post("/", router.handler.GetMetricJSONHandler)
		r.Get("/{type}/{name}", router.handler.GetMetricHandler)
	})
	r.Get("/ping", router.handler.PingHandler)
	r.Route("/updates", func(r chi.Router) {
		r.Post("/", router.handler.UpdateMetricsBatchHandler)
	})

	return r
}
