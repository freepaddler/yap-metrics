package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/freepaddler/yap-metrics/internal/logger"
	"github.com/freepaddler/yap-metrics/internal/models"
	"github.com/freepaddler/yap-metrics/internal/store"
)

var (
	l = &logger.L
)

type HTTPHandlers struct {
	storage store.Storage
}

func NewHTTPHandlers(s store.Storage) *HTTPHandlers {
	return &HTTPHandlers{
		storage: s,
	}
}

const (
	indexMetricHeader = `
	<html><head><title>Metrics Index</title></head>
	<body>
		<h2>Metrics Index</h2>
		<table border=1>
		<tr><th>Name</th><th>Type</th><th>Value</th></tr>
	`
)

// UpdateMetricHandler validates update request and writes metrics to storage
func (h *HTTPHandlers) UpdateMetricHandler(w http.ResponseWriter, r *http.Request) {
	l.Debug().Msgf("UpdateMetricHandler: Request received  URL=%v", r.URL)
	t := chi.URLParam(r, "type")  // metric type
	n := chi.URLParam(r, "name")  // metric name
	v := chi.URLParam(r, "value") // metric value
	switch t {
	case models.Counter:
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			l.Debug().Msgf("UpdateMetricHandler: wrong counter increment '%s'", v)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		h.storage.IncCounter(n, i)
	case models.Gauge:
		g, err := strconv.ParseFloat(v, 64)
		if err != nil {
			l.Debug().Msgf("UpdateMetricHandler: wrong gauge value '%s'", v)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		h.storage.SetGauge(n, g)
	default:
		l.Debug().Msgf("UpdateMetricHandler: wrong metric type '%s'", t)
		w.WriteHeader(http.StatusBadRequest)
	}
	w.WriteHeader(http.StatusOK)
}

// GetMetricHandler returns stored metrics
func (h *HTTPHandlers) GetMetricHandler(w http.ResponseWriter, r *http.Request) {
	l.Debug().Msgf("GetMetricHandler: Request received  URL=%v", r.URL)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	t := chi.URLParam(r, "type") // metric type
	n := chi.URLParam(r, "name") // metric name
	switch t {
	case models.Counter:
		if v, ok := h.storage.GetCounter(n); ok {
			w.Write([]byte(strconv.FormatInt(v, 10)))
			return
		}
	case models.Gauge:
		if v, ok := h.storage.GetGauge(n); ok {
			w.Write([]byte(strconv.FormatFloat(v, 'f', -1, 64)))
			return
		}
	default:
		l.Debug().Msgf("GetMetricHandler: bad metric type %s", t)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	l.Debug().Msgf("GetMetricHandler: requested metric %s does not exist", n)
	w.WriteHeader(http.StatusNotFound)
}

// IndexMetricHandler returns page with all metrics
func (h *HTTPHandlers) IndexMetricHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// TODO: use html templates
	w.Write([]byte(indexMetricHeader))
	for _, m := range h.storage.GetAllMetrics() {
		var val string
		switch m.Type {
		case models.Counter:
			val = strconv.FormatInt(m.IValue, 10)
		case models.Gauge:
			val = strconv.FormatFloat(m.FValue, 'f', -1, 64)
		default:
			continue
		}
		w.Write([]byte(fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%s</td></tr>", m.Name, m.Type, val)))
	}
	footer := fmt.Sprintf(`
		</table>
		<p><i>Updated: %s</i></p>
	</body>
	</html>
	`, time.Now().Format(time.UnixDate))
	w.Write([]byte(footer))
}
