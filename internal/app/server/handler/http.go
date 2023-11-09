// Package handler implements metrics server HTTP API
package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
	"github.com/freepaddler/yap-metrics/internal/pkg/models"
	"github.com/freepaddler/yap-metrics/internal/pkg/store"
)

const indexTmpl = `
<html><head><title>Metrics Index</title></head>
<body>
	<h2>Metrics Index</h2>
	<table border=1>
	<tr><th>Name</th><th>Type</th><th>Value</th></tr>
	{{ range . }}
	<tr>
		<td>{{ .Name }}</td>
		<td>{{ .Type }}</td>
		<td>{{ value . }}</td>
	</tr>
	{{ end }}
	</table>
	<p><i>Updated at {{ now }}</i></p>
</body>
</html>
`

type HTTPHandlers struct {
	storage store.Storage           // memory storage
	pStore  store.PersistentStorage // persistent storage (file or db)
}

// NewHTTPHandlers is HTTPHandlers constructor
func NewHTTPHandlers(s store.Storage) *HTTPHandlers {
	return &HTTPHandlers{
		storage: s,
		pStore:  nil,
	}
}

// SetPStorage updates persistent storage for HTTPHandlers
func (h *HTTPHandlers) SetPStorage(ps store.PersistentStorage) {
	h.pStore = ps
}

// IndexMetricHandler returns webpage with all metrics
//
//	curl -i http://localhost:8080/
func (h *HTTPHandlers) IndexMetricHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	funcMap := template.FuncMap{
		"value": func(m models.Metrics) (s string) {
			switch m.Type {
			case models.Counter:
				return strconv.FormatInt(*m.IValue, 10)
			case models.Gauge:
				return strconv.FormatFloat(*m.FValue, 'f', -1, 64)
			default:
				return
			}
		},
		"now": func() string { return time.Now().Format(time.UnixDate) },
	}
	tmpl, err := template.New("index").Funcs(funcMap).Parse(indexTmpl)
	if err != nil {
		logger.Log().Err(err).Msg("unable to parse index template")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	set := h.storage.Snapshot(false)
	sort.Slice(set, func(i, j int) bool {
		return set[i].Name < set[j].Name
	})
	err = tmpl.Execute(w, set)
	if err != nil {
		logger.Log().Err(err).Msg("unable to exec index template")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// GetMetricHandler returns requested metric value in plain text.
//
// # Responses
//   - 200/OK and value in plain text if metric found
//   - 400/BadRequest if request is invalid
//   - 404/NotFound if metric does not exist on server
//
// # Example
//
//	curl -i http://localhost:8080/value/counter/c1
func (h *HTTPHandlers) GetMetricHandler(w http.ResponseWriter, r *http.Request) {
	logger.Log().Debug().Msgf("GetMetricHandler: Request received  URL=%v", r.URL)
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
		logger.Log().Debug().Msgf("GetMetricHandler: bad metric type %s", t)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

// GetMetricJSONHandler returns requested metric in JSON.
//
// # Responses
//   - 200/OK and value in plain text if metric found
//   - 400/BadRequest if request is invalid
//   - 404/NotFound if metric does not exist on server
//   - 500/InternalServerError if any other error occurred
//
// # Example
//
//	curl -X POST -i http://localhost:8080/value -d '{"id":"g1","type":"gauge"}'
func (h *HTTPHandlers) GetMetricJSONHandler(w http.ResponseWriter, r *http.Request) {
	var m models.Metrics
	logger.Log().Debug().Msg("GetMetricJSONHandler: Request received: POST /value")
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		logger.Log().Warn().Err(err).Msg("GetMetricJSONHandler: unable to parse request JSON")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	code, ok := h.getMetric(&m)
	if ok {
		res, err := json.MarshalIndent(&m, "", "  ")
		if err != nil {
			logger.Log().Warn().Err(err).Msg("GetMetricJSONHandler: unable to marshal response JSON")
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
		logger.Log().Warn().Err(err).Msg("getMetricHTTP: invalid request")
		return http.StatusBadRequest, false
	}
	logger.Log().Debug().Msgf("getMetricHTTP: requested metric %+v", m)
	found, err := h.storage.GetMetric(m)
	if err != nil {
		logger.Log().Warn().Err(err).Msgf("getMetricHTTP: unable to get metric '%s' of type '%s'", m.Name, m.Type)
		return http.StatusBadRequest, false
	}
	if !found {
		logger.Log().Debug().Msgf("getMetricHTTP: no such metric '%s' of type '%s'", m.Name, m.Type)
		return http.StatusNotFound, false
	}
	return http.StatusOK, true
}

// UpdateMetricHandler creates new metric with value or updates value of existing metric.
// Single metric is passed in url path
//
// # Responses
//   - 200/OK on successful update
//   - 400/BadRequest if request is invalid
//
// # Example
//
//	curl -X POST -i http://localhost:8080/update/gauge/g1/-1.75
func (h *HTTPHandlers) UpdateMetricHandler(w http.ResponseWriter, r *http.Request) {
	logger.Log().Debug().Msgf("UpdateMetricHandler: Request received  URL=%v", r.URL)
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
			logger.Log().Debug().Msgf("UpdateMetricHandler: wrong counter increment '%s'", v)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		m.IValue = &i
	case models.Gauge:
		g, err := strconv.ParseFloat(v, 64)
		if err != nil {
			logger.Log().Debug().Msgf("UpdateMetricHandler: wrong gauge value '%s'", v)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		m.FValue = &g
	default:
		logger.Log().Debug().Msgf("UpdateMetricHandler: wrong metric type '%s'", t)
		w.WriteHeader(http.StatusBadRequest)
	}
	code, _ := h.updateMetric(&m)
	w.WriteHeader(code)
}

// UpdateMetricJSONHandler creates new metric with value or updates value of existing metric.
// Single metric is passed in JSON request body
//
// # Responses
//   - 200/OK on successful update, metric as JSON in body
//   - 400/BadRequest if request is invalid
//   - 500/InternalServerError if any other error occurred
//
// # Example
//
//	curl -X POST -i http://localhost:8080/update -d '{"id":"g2","type":"gauge","value":-1.75}'
func (h *HTTPHandlers) UpdateMetricJSONHandler(w http.ResponseWriter, r *http.Request) {
	var m models.Metrics
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		logger.Log().Warn().Err(err).Msg("UpdateMetricJSONHandler: unable to parse request JSON")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	code, ok := h.updateMetric(&m)
	if ok {
		res, err := json.MarshalIndent(&m, "", "  ")
		if err != nil {
			logger.Log().Warn().Err(err).Msg("UpdateMetricJSONHandler: unable to marshal response JSON")
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(res)
	}
	w.WriteHeader(code)
}

// UpdateMetricsBatchHandler creates new metrics with values or updates values of existing metrics.
// Multiple metrics are passed as JSON array in request body
//
// # Responses
//   - 200/OK
//   - 400/BadRequest if request is invalid
//   - 500/InternalServerError if any other error occurred
//
// # Example
//
//	curl -X POST -i http://localhost:8080/update -d '[{"id":"c101","type":"counter","delta":1},{"id":"g101","type":"gauge","value":-0.2}]'
func (h *HTTPHandlers) UpdateMetricsBatchHandler(w http.ResponseWriter, r *http.Request) {
	logger.Log().Debug().Msg("UpdateMetricsBatchHandler: request received")
	metrics := make([]models.Metrics, 0)
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		logger.Log().Warn().Err(err).Msg("UpdateMetricsBatchHandler: unable to parse request JSON")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	logger.Log().Debug().Msgf("Batch for update is: %v", metrics)
	for i := 0; i < len(metrics); i++ {
		err := validateMetric(&metrics[i])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	h.storage.UpdateMetrics(metrics, false)
	w.WriteHeader(http.StatusOK)
}

// updateMetric is a helper, that contains similar logic for UpdateMetricHandler and UpdateMetricJSONHandler
func (h *HTTPHandlers) updateMetric(m *models.Metrics) (int, bool) {
	if err := validateMetric(m); err != nil {
		logger.Log().Warn().Err(err).Msg("updateMetricHTTP: invalid request")
		return http.StatusBadRequest, false
	}
	logger.Log().Debug().Msgf("updateMetricHTTP: requested update of metric %+v", m)
	if (m.Type == models.Gauge && m.FValue == nil) ||
		(m.Type == models.Counter && m.IValue == nil) {
		logger.Log().Warn().Msgf("updateMetricHTTP: missing value for metric '%s' of type '%s'", m.Name, m.Type)
		return http.StatusBadRequest, false
	}
	h.storage.UpdateMetrics([]models.Metrics{*m}, false)
	return http.StatusOK, true
}

// PingHandler sends connectivity check to persistent storage
//
// # Responses
//   - 200/OK if persistent storage exists and accessible
//   - 500/InternalServerError otherwise
func (h *HTTPHandlers) PingHandler(w http.ResponseWriter, r *http.Request) {
	logger.Log().Debug().Msgf("PingHandler: Request received  URL=%v", r.URL)
	if h.pStore == nil {
		logger.Log().Debug().Msg("PingHandler: no persistent storage is setup")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := h.pStore.Ping(); err != nil {
		logger.Log().Warn().Err(err).Msg("PingHandler: persistent storage connection failed")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// validateMetric check json unmarshalled metric validity. Returns error if metric is invalid
func validateMetric(m *models.Metrics) (err error) {
	if m.Name == "" {
		err = errors.New("missing metric name")
	}
	if m.Type != models.Gauge && m.Type != models.Counter {
		err = fmt.Errorf("invalid metric '%s' type '%s", m.Name, m.Type)
	}
	return err
}
