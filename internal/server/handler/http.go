package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/freepaddler/yap-metrics/internal/logger"
	"github.com/freepaddler/yap-metrics/internal/models"
	"github.com/freepaddler/yap-metrics/internal/store"
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
	logger.Log.Debug().Msgf("UpdateMetricHandler: Request received  URL=%v", r.URL)
	t := chi.URLParam(r, "type")  // metric type
	n := chi.URLParam(r, "name")  // metric name
	v := chi.URLParam(r, "value") // metric value
	switch t {
	case models.Counter:
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			logger.Log.Debug().Msgf("UpdateMetricHandler: wrong counter increment '%s'", v)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		h.storage.IncCounter(n, i)
	case models.Gauge:
		g, err := strconv.ParseFloat(v, 64)
		if err != nil {
			logger.Log.Debug().Msgf("UpdateMetricHandler: wrong gauge value '%s'", v)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		h.storage.SetGauge(n, g)
	default:
		logger.Log.Debug().Msgf("UpdateMetricHandler: wrong metric type '%s'", t)
		w.WriteHeader(http.StatusBadRequest)
	}
	w.WriteHeader(http.StatusOK)
}

// GetMetricHandler returns stored metrics
func (h *HTTPHandlers) GetMetricHandler(w http.ResponseWriter, r *http.Request) {
	logger.Log.Debug().Msgf("GetMetricHandler: Request received  URL=%v", r.URL)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	t := chi.URLParam(r, "type") // metric type
	n := chi.URLParam(r, "name") // metric name
	switch t {
	case models.Counter:
		if v, ok := h.storage.GetCounter(n); ok {
			w.Write([]byte(strconv.FormatInt(*v, 10)))
			return
		}
	case models.Gauge:
		if v, ok := h.storage.GetGauge(n); ok {
			w.Write([]byte(strconv.FormatFloat(*v, 'f', -1, 64)))
			return
		}
	default:
		logger.Log.Debug().Msgf("GetMetricHandler: bad metric type %s", t)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	logger.Log.Debug().Msgf("GetMetricHandler: requested metric %s does not exist", n)
	w.WriteHeader(http.StatusNotFound)
}

// IndexMetricHandler returns page with all metrics
func (h *HTTPHandlers) IndexMetricHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// TODO: use html templates
	w.Write([]byte(indexMetricHeader))
	for _, m := range h.storage.Snapshot() {
		var val string
		switch m.Type {
		case models.Counter:
			val = strconv.FormatInt(*m.IValue, 10)
		case models.Gauge:
			val = strconv.FormatFloat(*m.FValue, 'f', -1, 64)
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

func (h *HTTPHandlers) GetMetricJSONHandler(w http.ResponseWriter, r *http.Request) {
	var m models.Metrics
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		logger.Log.Warn().Err(err).Msg("GetMetricJSONHandler: unable to parse request JSON")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := checkMetric(&m); err != nil {
		logger.Log.Warn().Err(err).Msg("GetMetricJSONHandler: invalid request")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	found, err := h.storage.GetMetric(&m)
	if err != nil {
		logger.Log.Warn().Err(err).Msgf("GetMetricJSONHandler: unable to get metric '%s' of type '%s'", m.Name, m.Type)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !found {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	res, err := json.MarshalIndent(&m, "", "  ")
	if err != nil {
		logger.Log.Warn().Err(err).Msg("GetMetricJSONHandler: unable to marshal response JSON")
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}

func (h *HTTPHandlers) UpdateMetricJSONHandler(w http.ResponseWriter, r *http.Request) {
	var m models.Metrics
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		logger.Log.Warn().Err(err).Msg("UpdateMetricJSONHandler: unable to parse request JSON")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := checkMetric(&m); err != nil {
		logger.Log.Warn().Err(err).Msg("GetMetricJSONHandler: invalid request")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// check that value exists
	// TODO: may be using validator is better idea
	if (m.Type == models.Gauge && m.FValue == nil) ||
		(m.Type == models.Counter && m.IValue == nil) {
		logger.Log.Warn().Msgf("UpdateMetricJSONHandler: missing value for metric '%s' of type '%s'", m.Name, m.Type)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.storage.SetMetric(&m); err != nil {
		logger.Log.Warn().Err(err).Msg("UpdateMetricJSONHandler: failed to update metric")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	res, err := json.MarshalIndent(&m, "", "  ")
	if err != nil {
		logger.Log.Warn().Err(err).Msg("UpdateMetricJSONHandler: unable to marshal response JSON")
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
}

func checkMetric(m *models.Metrics) (err error) {
	if m.Name == "" {
		err = errors.New("missing metric name")
	}
	if m.Type != models.Gauge && m.Type != models.Counter {
		err = fmt.Errorf("invalid metric '%s' type '%s", m.Name, m.Type)
	}
	return err
}
