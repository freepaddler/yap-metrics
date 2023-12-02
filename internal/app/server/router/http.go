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
	checkIp      Middleware
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

func WithIpMatcher(mw Middleware) func(router *Router) {
	return func(router *Router) {
		router.checkIp = mw
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

	// middleware order is important
	if router.log != nil {
		r.Use(router.log)
	}
	if router.gunzip != nil {
		r.Use(router.gunzip)
	}
	if router.gzip != nil {
		r.Use(router.gzip)
	}

	if router.profilerPath != "" {
		r.Mount(router.profilerPath, middleware.Profiler())
	}

	r.Get("/", router.handler.IndexMetricHandler)
	r.Route("/value", func(r chi.Router) {
		r.Post("/", router.handler.GetMetricJSONHandler)
		r.Get("/{type}/{name}", router.handler.GetMetricHandler)
	})
	r.Get("/ping", router.handler.PingHandler)

	// only update routes
	r.Group(func(r chi.Router) {
		if router.checkIp != nil {
			r.Use(router.checkIp)
		}
		if router.crypt != nil {
			r.Use(router.crypt)
		}
		if router.sign != nil {
			r.Use(router.sign)
		}
		r.Route("/update", func(r chi.Router) {
			r.Post("/", router.handler.UpdateMetricJSONHandler)
			r.Post("/{type}/{name}/{value}", router.handler.UpdateMetricHandler)
		})
		r.Route("/updates", func(r chi.Router) {
			r.Post("/", router.handler.UpdateMetricsBatchHandler)
		})
	})

	return r
}
