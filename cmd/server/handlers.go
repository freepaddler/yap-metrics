package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/freepaddler/yap-metrics/internal/models"
)

// UpdateHandler validates update request and writes metrics to storage
func (srv *MetricsServer) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("UpdateHandler: Request received  URL=%v\n", r.URL)
	switch chi.URLParam(r, "type") {
	case models.Counter:
		v, err := strconv.ParseInt(chi.URLParam(r, "value"), 10, 64)
		if err != nil {
			fmt.Printf("UpdateHandler: wrong counter increment '%s'\n", chi.URLParam(r, "value"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		srv.storage.CounterSet(chi.URLParam(r, "name"), v)
	case models.Gauge:
		v, err := strconv.ParseFloat(chi.URLParam(r, "value"), 64)
		if err != nil {
			fmt.Printf("UpdateHandler: wrong gauge value '%s'\n", chi.URLParam(r, "value"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		srv.storage.GaugeSet(chi.URLParam(r, "name"), v)
	default:
		fmt.Printf("UpdateHandler: wrong metric type '%s'\n", chi.URLParam(r, "type"))
		w.WriteHeader(http.StatusBadRequest)
	}
}

// ValueHandler returns stored metrics
func (srv *MetricsServer) ValueHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("ValueHandler: Request received  URL=%v\n", r.URL)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	switch t := chi.URLParam(r, "type"); t {
	case models.Counter:
		if v, ok := srv.storage.CounterGet(chi.URLParam(r, "name")); ok {
			w.Write([]byte(strconv.FormatInt(v, 10)))
			return
		}
	case models.Gauge:
		if v, ok := srv.storage.GaugeGet(chi.URLParam(r, "name")); ok {
			w.Write([]byte(strconv.FormatFloat(v, 'f', -1, 64)))
			return
		}
	default:
		fmt.Printf("ValueHandler: bad metric type %s\n", chi.URLParam(r, "type"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fmt.Printf("ValueHandler: requested metric %s does not exist\n", chi.URLParam(r, "name"))
	w.WriteHeader(http.StatusNotFound)
}

// IndexHandler returns page with all metrics
func (srv *MetricsServer) IndexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	header := `
	<html><head><title>Metrics Index</title></head>
	<body>
		<h2>Metrics Index</h2>
		<table border=1>
		<tr><th>Name</th><th>Type</th><th>Value</th></tr>
	`
	w.Write([]byte(header))
	for _, m := range srv.storage.GetMetrics() {
		var val string
		switch m.Type {
		case models.Counter:
			//val = fmt.Sprintf("%d", m.Value)
			val = strconv.FormatInt(m.Value, 10)
		case models.Gauge:
			//val = fmt.Sprintf("%.3f", m.Gauge)
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
