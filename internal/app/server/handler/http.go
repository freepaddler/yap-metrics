package handler

import (
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
	pStore  store.PersistentStorage
}

func NewHTTPHandlers(s store.Storage) *HTTPHandlers {
	return &HTTPHandlers{
		storage: s,
		pStore:  nil,
	}
}

func (h *HTTPHandlers) SetPStorage(ps store.PersistentStorage) {
	h.pStore = ps
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

// IndexMetricHandler returns webpage page with all metrics
func (h *HTTPHandlers) IndexMetricHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// TODO: use html templates
	w.Write([]byte(indexMetricHeader))
	for _, m := range h.storage.Snapshot(false) {
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
		<p><i>Updated at %s</i></p>
	</body>
	</html>
	`, time.Now().Format(time.UnixDate))
	w.Write([]byte(footer))
}

// GetMetricHandler returns stored metrics for url-param request
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

// GetMetricJSONHandler returns stored metrics for JSON single-metric request
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

// getMetric is a helper, that contains similar logic for GetMetricHandler and GetMetricJSONHandler
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

// UpdateMetricHandler validates url-params update request and writes metrics to storage
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

// UpdateMetricJSONHandler validates JSON single-metric update request and writes metrics to storage
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

func (h *HTTPHandlers) UpdateMetricsBatchHandler(w http.ResponseWriter, r *http.Request) {
	logger.Log.Debug().Msg("UpdateMetricsBatchHandler: request received")
	metrics := make([]models.Metrics, 0)
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		logger.Log.Warn().Err(err).Msg("UpdateMetricsBatchHandler: unable to parse request JSON")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	logger.Log.Debug().Msgf("Batch for update is: %v", metrics)
	//reqLen := len(metrics)
	// validate parsed metrics
	//invalid := make([]models.Metrics, 0)
	for i := 0; i < len(metrics); i++ {
		err := validateMetric(&metrics[i])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
			//invalid = append(invalid, metrics[i])
			//metrics[i] = metrics[len(metrics)-1]
			//metrics = metrics[:len(metrics)-1]
			//i--
			//continue
		}
	}
	invalid2 := h.storage.UpdateMetrics(metrics, false)
	if len(invalid2) > 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
		//invalid = append(invalid, invalid2...)
	}
	//if len(invalid) == reqLen {
	//	logger.Log.Warn().Msg("UpdateMetricsBatchHandler: all metrics are invalid")
	//	w.WriteHeader(http.StatusBadRequest)
	//	return
	//}
	//res, err := json.MarshalIndent(&metrics, "", "  ")
	//if err != nil {
	//	logger.Log.Warn().Msg("UpdateMetricsBatchHandler: unable to marshal response JSON")
	//	w.WriteHeader(http.StatusInternalServerError)
	//	return
	//}
	//logger.Log.Debug().Msgf("Marshalled request: %s", res)
	//w.Header().Set("Content-Type", "application/json")
	//w.Write(res)
	w.WriteHeader(http.StatusOK)
}

// updateMetric is a helper, that contains similar logic for UpdateMetricHandler and UpdateMetricJSONHandler
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
	if invalid := h.storage.UpdateMetrics([]models.Metrics{*m}, false); len(invalid) > 0 {
		logger.Log.Warn().Msgf("updateMetricHTTP: failed to update metric %+v", m)
		return http.StatusInternalServerError, false
	}
	return http.StatusOK, true
}

// PingHandler sends connectivity check to persistent storage
func (h *HTTPHandlers) PingHandler(w http.ResponseWriter, r *http.Request) {
	logger.Log.Debug().Msgf("PingHandler: Request received  URL=%v", r.URL)
	if h.pStore == nil {
		logger.Log.Debug().Msg("PingHandler: no persistent storage is setup")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := h.pStore.Ping(); err != nil {
		logger.Log.Warn().Err(err).Msg("PingHandler: persistent storage connection failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// validateMetric check json unmarshalled metric validity
func validateMetric(m *models.Metrics) (err error) {
	if m.Name == "" {
		err = errors.New("missing metric name")
	}
	if m.Type != models.Gauge && m.Type != models.Counter {
		err = fmt.Errorf("invalid metric '%s' type '%s", m.Name, m.Type)
	}
	return err
}
