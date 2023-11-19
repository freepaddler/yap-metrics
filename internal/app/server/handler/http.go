// Package handler implements metrics server HTTP API
package handler

import (
	"encoding/json"
	"errors"
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

//go:generate mockgen -source $GOFILE -package=mocks -destination ../../../../mocks/HTTPHandlerStorage_mock.go

type HTTPHandlerStorage interface {
	GetAll() []models.Metrics
	GetOne(request models.MetricRequest) (models.Metrics, error)
	UpdateOne(metric *models.Metrics) error
	UpdateMany(metrics []models.Metrics) error
	Ping() error
}

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
	storage HTTPHandlerStorage // server handler methods
}

// NewHTTPHandlers is HTTPHandlers constructor
func NewHTTPHandlers(storage HTTPHandlerStorage) *HTTPHandlers {
	return &HTTPHandlers{
		storage: storage,
	}
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
	set := h.storage.GetAll()
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
	req, err := models.NewMetricRequest(chi.URLParam(r, "name"), chi.URLParam(r, "type"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	m, err := h.storage.GetOne(req)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrInvalidMetric):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, store.ErrMetricNotFound):
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(m.StringVal()))
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
	var req models.MetricRequest
	logger.Log().Debug().Msg("GetMetricJSONHandler: Request received: POST /value")
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Log().Warn().Err(err).Msg("GetMetricJSONHandler: unable to parse request JSON")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	m, err := h.storage.GetOne(req)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrInvalidMetric):
			w.WriteHeader(http.StatusBadRequest)
		case errors.Is(err, store.ErrMetricNotFound):
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	res, err := json.MarshalIndent(&m, "", "  ")
	if err != nil {
		logger.Log().Warn().Err(err).Msg("GetMetricJSONHandler: unable to marshal response JSON")
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(res)

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
	m, err := models.NewMetric(chi.URLParam(r, "name"), chi.URLParam(r, "type"), chi.URLParam(r, "value"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = h.storage.UpdateOne(&m)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Write([]byte(m.StringVal()))

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
	err := h.storage.UpdateOne(&m)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	res, err := json.MarshalIndent(&m, "", "  ")
	if err != nil {
		logger.Log().Warn().Err(err).Msg("UpdateMetricJSONHandler: unable to marshal response JSON")
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(res)
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
	if len(metrics) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err := h.storage.UpdateMany(metrics)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// TODO: after persistent storage

// PingHandler sends connectivity check to persistent storage
//
// # Responses
//   - 200/OK if persistent storage exists and accessible
//   - 500/InternalServerError otherwise
func (h *HTTPHandlers) PingHandler(w http.ResponseWriter, r *http.Request) {
	logger.Log().Debug().Msgf("PingHandler: Request received  URL=%v", r.URL)
	if err := h.storage.Ping(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
