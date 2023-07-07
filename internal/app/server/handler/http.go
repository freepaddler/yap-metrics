package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
	"github.com/freepaddler/yap-metrics/internal/pkg/models"
	"github.com/freepaddler/yap-metrics/internal/pkg/store"
)

type HTTPHandlers struct {
	storage store.Storage
	db      *sql.DB
}

func NewHTTPHandlers(s store.Storage, db *sql.DB) *HTTPHandlers {
	return &HTTPHandlers{
		storage: s,
		db:      db,
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
		<p><i>SaveMetric: %s</i></p>
	</body>
	</html>
	`, time.Now().Format(time.UnixDate))
	w.Write([]byte(footer))
}

// GetMetricHandler returns stored metrics
func (h *HTTPHandlers) GetMetricHandler(w http.ResponseWriter, r *http.Request) {
	logger.Log.Debug().Msgf("GetMetricHandler: Request received  URL=%v", r.URL)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	t := chi.URLParam(r, "type") // metric type
	n := chi.URLParam(r, "name") // metric name
	m := models.Metrics{
		Name: n,
		Type: t,
	}
	code, _ := h.getMetric(&m)
	w.WriteHeader(code)
	switch m.Type {
	case models.Counter:
		w.Write([]byte(strconv.FormatInt(*m.IValue, 10)))
		return
	case models.Gauge:
		w.Write([]byte(strconv.FormatFloat(*m.FValue, 'f', -1, 64)))
		return
	default:
		logger.Log.Debug().Msgf("GetMetricHandler: bad metric type %s", t)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (h *HTTPHandlers) GetMetricJSONHandler(w http.ResponseWriter, r *http.Request) {
	var m models.Metrics
	logger.Log.Debug().Msg("GetMetricJSONHandler: Request received: POST /value")
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		logger.Log.Warn().Err(err).Msg("GetMetricJSONHandler: unable to parse request JSON")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	code, ok := h.getMetric(&m)
	if ok {
		res, err := json.MarshalIndent(&m, "", "  ")
		if err != nil {
			logger.Log.Warn().Err(err).Msg("GetMetricJSONHandler: unable to marshal response JSON")
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(res)
	}
	w.WriteHeader(code)
}

func (h *HTTPHandlers) getMetric(m *models.Metrics) (int, bool) {
	if err := validateMetric(m); err != nil {
		logger.Log.Warn().Err(err).Msg("getMetricHTTP: invalid request")
		return http.StatusBadRequest, false
	}
	logger.Log.Debug().Msgf("getMetricHTTP: requested metric %+v", m)
	found, err := h.storage.GetMetric(m)
	if err != nil {
		logger.Log.Warn().Err(err).Msgf("getMetricHTTP: unable to get metric '%s' of type '%s'", m.Name, m.Type)
		return http.StatusBadRequest, false
	}
	if !found {
		logger.Log.Debug().Msgf("getMetricHTTP: no such metric '%s' of type '%s'", m.Name, m.Type)
		return http.StatusNotFound, false
	}
	return http.StatusOK, true
}

// UpdateMetricHandler validates update request and writes metrics to storage
func (h *HTTPHandlers) UpdateMetricHandler(w http.ResponseWriter, r *http.Request) {
	logger.Log.Debug().Msgf("UpdateMetricHandler: Request received  URL=%v", r.URL)
	t := chi.URLParam(r, "type")  // metric type
	n := chi.URLParam(r, "name")  // metric name
	v := chi.URLParam(r, "value") // metric value
	m := models.Metrics{
		Name: n,
		Type: t,
	}
	switch t {
	case models.Counter:
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			logger.Log.Debug().Msgf("UpdateMetricHandler: wrong counter increment '%s'", v)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		m.IValue = &i
	case models.Gauge:
		g, err := strconv.ParseFloat(v, 64)
		if err != nil {
			logger.Log.Debug().Msgf("UpdateMetricHandler: wrong gauge value '%s'", v)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		m.FValue = &g
	default:
		logger.Log.Debug().Msgf("UpdateMetricHandler: wrong metric type '%s'", t)
		w.WriteHeader(http.StatusBadRequest)
	}
	code, _ := h.updateMetric(&m)
	w.WriteHeader(code)
}

func (h *HTTPHandlers) UpdateMetricJSONHandler(w http.ResponseWriter, r *http.Request) {
	var m models.Metrics
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		logger.Log.Warn().Err(err).Msg("UpdateMetricJSONHandler: unable to parse request JSON")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	code, ok := h.updateMetric(&m)
	if ok {
		res, err := json.MarshalIndent(&m, "", "  ")
		if err != nil {
			logger.Log.Warn().Err(err).Msg("UpdateMetricJSONHandler: unable to marshal response JSON")
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(res)
	}
	w.WriteHeader(code)
}

func (h *HTTPHandlers) updateMetric(m *models.Metrics) (int, bool) {
	if err := validateMetric(m); err != nil {
		logger.Log.Warn().Err(err).Msg("updateMetricHTTP: invalid request")
		return http.StatusBadRequest, false
	}
	logger.Log.Debug().Msgf("updateMetricHTTP: requested update of metric %+v", m)
	if (m.Type == models.Gauge && m.FValue == nil) ||
		(m.Type == models.Counter && m.IValue == nil) {
		logger.Log.Warn().Msgf("updateMetricHTTP: missing value for metric '%s' of type '%s'", m.Name, m.Type)
		return http.StatusBadRequest, false
	}
	if err := h.storage.SetMetric(m); err != nil {
		logger.Log.Warn().Err(err).Msg("updateMetricHTTP: failed to update metric")
		return http.StatusInternalServerError, false
	}
	return http.StatusOK, true
}

func (h *HTTPHandlers) PingDBHandler(w http.ResponseWriter, r *http.Request) {
	logger.Log.Debug().Msgf("PingDBHandler: Request received  URL=%v", r.URL)
	if h.db == nil {
		logger.Log.Debug().Msg("PingDBHandler: no database connection setup")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// avoid wrong hostnames
	ctxPing, ctxPingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer ctxPingCancel()
	if err := h.db.PingContext(ctxPing); err != nil {
		logger.Log.Warn().Err(err).Msg("PingDBHandler: database connection failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func validateMetric(m *models.Metrics) (err error) {
	if m.Name == "" {
		err = errors.New("missing metric name")
	}
	if m.Type != models.Gauge && m.Type != models.Counter {
		err = fmt.Errorf("invalid metric '%s' type '%s", m.Name, m.Type)
	}
	return err
}
