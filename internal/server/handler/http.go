package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/freepaddler/yap-metrics/internal/models"
	"github.com/freepaddler/yap-metrics/internal/store"
)

type HttpHandlers struct {
	storage store.Storage
}

func NewHttpHandlers(srv store.Storage) *HttpHandlers {
	return &HttpHandlers{
		storage: srv,
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
func (h *HttpHandlers) UpdateMetricHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("UpdateMetricHandler: Request received  URL=%v\n", r.URL)
	t := chi.URLParam(r, "type")  // metric type
	n := chi.URLParam(r, "name")  // metric name
	v := chi.URLParam(r, "value") // metric value
	switch t {
	case models.Counter:
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			fmt.Printf("UpdateMetricHandler: wrong counter increment '%s'\n", v)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		h.storage.IncCounter(n, i)
	case models.Gauge:
		g, err := strconv.ParseFloat(v, 64)
		if err != nil {
			fmt.Printf("UpdateMetricHandler: wrong gauge value '%s'\n", v)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		h.storage.SetGauge(n, g)
	default:
		fmt.Printf("UpdateMetricHandler: wrong metric type '%s'\n", t)
		w.WriteHeader(http.StatusBadRequest)
	}
}

// GetMetricHandler returns stored metrics
func (h *HttpHandlers) GetMetricHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("GetMetricHandler: Request received  URL=%v\n", r.URL)
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
		fmt.Printf("GetMetricHandler: bad metric type %s\n", t)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Printf("GetMetricHandler: requested metric %s does not exist\n", n)
	w.WriteHeader(http.StatusNotFound)
}

// IndexMetricHandler returns page with all metrics
func (h *HttpHandlers) IndexMetricHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// TODO: use html templates
	w.Write([]byte(indexMetricHeader))
	for _, m := range h.storage.GetAllMetrics() {
		var val string
		switch m.Type {
		case models.Counter:
			val = strconv.FormatInt(m.Value, 10)
		case models.Gauge:
			val = strconv.FormatFloat(m.Gauge, 'f', -1, 64)
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
