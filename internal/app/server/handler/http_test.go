package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/freepaddler/yap-metrics/internal/pkg/models"
	"github.com/freepaddler/yap-metrics/internal/pkg/store"
	"github.com/freepaddler/yap-metrics/internal/pkg/store/file"
	"github.com/freepaddler/yap-metrics/internal/pkg/store/memory"
)

func TestHTTPHandlers_Index(t *testing.T) {
	s := memory.NewMemStorage()
	h := NewHTTPHandlers(s)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.IndexMetricHandler(w, req)
	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "text/html; charset=utf-8", res.Header.Get("Content-Type"))
}

func TestHTTPHandlers_UpdateMetric(t *testing.T) {
	s := memory.NewMemStorage()
	h := NewHTTPHandlers(s)
	tests := []struct {
		name   string
		code   int
		mType  string
		mName  string
		mValue string
	}{
		{
			name:   "success new counter",
			code:   http.StatusOK,
			mType:  models.Counter,
			mName:  "c1",
			mValue: "10",
		},
		{
			name:   "success update counter",
			code:   http.StatusOK,
			mType:  models.Counter,
			mName:  "c1",
			mValue: "10",
		},
		{
			name:   "success new gauge",
			code:   http.StatusOK,
			mType:  models.Gauge,
			mName:  "g1",
			mValue: "-1.75",
		},
		{
			name:   "success update gauge",
			code:   http.StatusOK,
			mType:  models.Gauge,
			mName:  "g1",
			mValue: "1",
		},
		{
			name:   "invalid metric type",
			code:   http.StatusBadRequest,
			mType:  "gauge1",
			mName:  "g1",
			mValue: "1",
		},
		{
			name:   "invalid counter value string",
			code:   http.StatusBadRequest,
			mType:  models.Counter,
			mName:  "c2",
			mValue: "none",
		},
		{
			name:   "invalid counter value float",
			code:   http.StatusBadRequest,
			mType:  models.Counter,
			mName:  "c2",
			mValue: "-0.117",
		},
		{
			name:   "invalid gauge value string",
			code:   http.StatusBadRequest,
			mType:  models.Gauge,
			mName:  "g1",
			mValue: "something",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/update/{type}/{name}/{value}", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("type", tt.mType)
			rctx.URLParams.Add("name", tt.mName)
			rctx.URLParams.Add("value", tt.mValue)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			w := httptest.NewRecorder()
			h.UpdateMetricHandler(w, req)
			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.code, res.StatusCode)
		})
	}
}

func TestHTTPHandlers_GetMetric(t *testing.T) {
	s := memory.NewMemStorage()
	h := NewHTTPHandlers(s)
	var cValue int64 = 10
	var cName = "c1"
	var gValue float64 = -0.110
	var gName = "g1"
	h.storage.IncCounter(cName, cValue)
	h.storage.SetGauge(gName, gValue)
	tests := []struct {
		name  string
		code  int
		mType string
		mName string
		want  string
	}{
		{
			name:  "success counter",
			code:  http.StatusOK,
			mType: models.Counter,
			mName: cName,
			//want:  fmt.Sprintf("%d", cValue),
			want: strconv.FormatInt(cValue, 10),
		},
		{
			name:  "success gauge",
			code:  http.StatusOK,
			mType: models.Gauge,
			mName: gName,
			//want:  fmt.Sprintf("%.3f", gValue),
			want: strconv.FormatFloat(gValue, 'f', -1, 64),
		},
		{
			name:  "invalid counter type",
			code:  http.StatusBadRequest,
			mType: "qqq",
			mName: gName,
		},
		{
			name:  "invalid counter name",
			code:  http.StatusNotFound,
			mType: models.Counter,
			mName: gName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/value/{type}/{name}", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("type", tt.mType)
			rctx.URLParams.Add("name", tt.mName)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			w := httptest.NewRecorder()
			h.GetMetricHandler(w, req)
			res := w.Result()
			defer res.Body.Close()

			require.Equal(t, tt.code, res.StatusCode)
			if res.StatusCode == http.StatusOK {
				resBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				assert.Equal(t, tt.want, string(resBody))
			}
		})
	}
}

func TestHTTPHandlers_GetMetricJSON(t *testing.T) {
	s := memory.NewMemStorage()
	h := NewHTTPHandlers(s)
	var cValue int64 = 10
	var cName = "c1"
	var gValue float64 = -0.110
	var gName = "g1"
	h.storage.IncCounter(cName, cValue)
	h.storage.SetGauge(gName, gValue)
	tests := []struct {
		name   string
		code   int
		mType  string
		mName  string
		cValue *int64
		gValue *float64
	}{
		{
			name:   "success counter",
			code:   200,
			mType:  models.Counter,
			mName:  cName,
			cValue: &cValue,
			gValue: nil,
		},
		{
			name:   "success gauge",
			code:   200,
			mType:  models.Gauge,
			mName:  gName,
			gValue: &gValue,
			cValue: nil,
		},
		{
			name:  "invalid counter type",
			code:  400,
			mType: "qqqqq",
			mName: "d1",
		},
		{
			name:  "invalid gauge name",
			code:  404,
			mType: models.Gauge,
			mName: "gauge101",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, _ := json.Marshal(models.Metrics{
				Name: tt.mName,
				Type: tt.mType,
			})
			buf := bytes.NewBuffer(m)
			req := httptest.NewRequest(http.MethodPost, "/value", buf)

			w := httptest.NewRecorder()
			h.GetMetricJSONHandler(w, req)
			res := w.Result()
			defer res.Body.Close()

			require.Equal(t, tt.code, res.StatusCode)
			if res.StatusCode == http.StatusOK {
				resBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				var resJSON models.Metrics
				require.NoError(t, json.Unmarshal(resBody, &resJSON))
				assert.Equal(t, models.Metrics{
					Name:   tt.mName,
					Type:   tt.mType,
					FValue: tt.gValue,
					IValue: tt.cValue,
				}, resJSON)
			}
		})
	}
}

func TestHTTPHandlers_UpdateMetricJSON(t *testing.T) {
	s := memory.NewMemStorage()
	h := NewHTTPHandlers(s)
	tests := []struct {
		name       string
		code       int
		reqString  string
		wantString string
	}{
		{
			name: "success new counter",
			code: http.StatusOK,
			reqString: fmt.Sprintf(`
				{ 
					"id":"c1",
					"type":"%s",
					"delta":10
				}`, models.Counter),
			wantString: fmt.Sprintf(`
				{ 
					"id":"c1",
					"type":"%s",
					"delta":10
				}`, models.Counter),
		},
		{
			name: "success update counter",
			code: http.StatusOK,
			reqString: fmt.Sprintf(`
				{ 
					"id":"c1",
					"type":"%s",
					"delta":10
				}`, models.Counter),
			wantString: fmt.Sprintf(`
				{ 
					"id":"c1",
					"type":"%s",
					"delta":20
				}`, models.Counter),
		},
		{
			name: "success new gauge",
			code: http.StatusOK,
			reqString: fmt.Sprintf(`
				{ 
					"id":"c1",
					"type":"%s",
					"value":-1.75
				}`, models.Gauge),
			wantString: fmt.Sprintf(`
				{ 
					"id":"c1",
					"type":"%s",
					"value":-1.75
				}`, models.Gauge),
		},
		{
			name: "success update gauge",
			code: http.StatusOK,
			reqString: fmt.Sprintf(`
				{ 
					"id":"c1",
					"type":"%s",
					"value":1.000
				}`, models.Gauge),
			wantString: fmt.Sprintf(`
				{ 
					"id":"c1",
					"type":"%s",
					"value":1
				}`, models.Gauge),
		},
		{
			name: "invalid metric type",
			code: http.StatusBadRequest,
			reqString: fmt.Sprintf(`
				{ 
					"id":"g1",
					"type":"%s",
					"value":1.000
				}`, "gauge1"),
		},
		{
			name: "invalid counter value string",
			code: http.StatusBadRequest,
			reqString: fmt.Sprintf(`
				{ 
					"id":"c1",
					"type":"%s",
					"value":"something"
				}`, models.Counter),
		},
		{
			name: "invalid counter value float",
			code: http.StatusBadRequest,
			reqString: fmt.Sprintf(`
				{ 
					"id":"c1",
					"type":"%s",
					"value":0.21
				}`, models.Counter),
		},
		{
			name: "invalid gauge value string",
			code: http.StatusBadRequest,
			reqString: fmt.Sprintf(`
				{ 
					"id":"g1",
					"type":"%s",
					"value":"something"
				}`, models.Gauge),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer([]byte(tt.reqString))
			req := httptest.NewRequest(http.MethodPost, "/update", buf)

			w := httptest.NewRecorder()
			h.UpdateMetricJSONHandler(w, req)
			res := w.Result()
			defer res.Body.Close()

			require.Equal(t, tt.code, res.StatusCode)
			if res.StatusCode == http.StatusOK {
				resBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				var resJSON models.Metrics
				require.NoError(t, json.Unmarshal(resBody, &resJSON))
				var m models.Metrics
				json.Unmarshal([]byte(tt.wantString), &m)
				assert.Equal(t, m, resJSON)
			}
		})
	}
}

func TestHTTPHandlers_UpdateMetricBatch(t *testing.T) {
	s := memory.NewMemStorage()
	h := NewHTTPHandlers(s)
	tests := []struct {
		name       string
		code       int
		reqString  string
		wantString string
	}{
		{
			name: "success new counter and gauge",
			code: http.StatusOK,
			reqString: fmt.Sprintf(`
				[{ 
					"id":"c1",
					"type":"%s",
					"delta":10
				},
				{ 
					"id":"g1",
					"type":"%s",
					"value":-1.75
				}]`,
				models.Counter, models.Gauge),
		},
		{
			name: "success update counter and gauge",
			code: http.StatusOK,
			reqString: fmt.Sprintf(`
				[{ 
					"id":"c1",
					"type":"%s",
					"delta":8
				},
				{ 
					"id":"g1",
					"type":"%s",
					"value":0.07
				}]`,
				models.Counter, models.Gauge),
		},
		{
			name: "invalid metric type",
			code: http.StatusBadRequest,
			reqString: fmt.Sprintf(`
				[{ 
					"id":"c1",
					"type":"%s",
					"delta":8
				},
				{ 
					"id":"g1",
					"type":"%s",
					"value":0.07
				}]`,
				models.Counter, "invalid"),
		},
		{
			name: "invalid counter value string",
			code: http.StatusBadRequest,
			reqString: fmt.Sprintf(`
				[{ 
					"id":"c1",
					"type":"%s",
					"delta":"string"
				},
				{ 
					"id":"g1",
					"type":"%s",
					"value":0.07
				}]`,
				models.Counter, models.Gauge),
		},
		{
			name: "invalid counter value float",
			code: http.StatusBadRequest,
			reqString: fmt.Sprintf(`
				[{ 
					"id":"c1",
					"type":"%s",
					"delta":8.88
				},
				{ 
					"id":"g1",
					"type":"%s",
					"value":0.07
				}]`,
				models.Counter, models.Gauge),
		},
		{
			name: "invalid gauge value string",
			code: http.StatusBadRequest,
			reqString: fmt.Sprintf(`
				[{ 
					"id":"c1",
					"type":"%s",
					"delta":8
				},
				{ 
					"id":"g1",
					"type":"%s",
					"value":"string"
				}]`,
				models.Counter, models.Gauge),
		},
		{
			name:      "invalid json",
			code:      http.StatusBadRequest,
			reqString: "[this is not a json}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer([]byte(tt.reqString))
			req := httptest.NewRequest(http.MethodPost, "/update", buf)

			w := httptest.NewRecorder()
			h.UpdateMetricsBatchHandler(w, req)
			res := w.Result()
			defer res.Body.Close()

			require.Equal(t, tt.code, res.StatusCode)
		})
	}
}

func Test_validateMetric(t *testing.T) {
	tests := []struct {
		name    string
		metric  models.Metrics
		wantErr bool
	}{
		{
			name: "correct metric",
			metric: models.Metrics{
				Name: "someMetric",
				Type: models.Gauge,
			},
			wantErr: false,
		},
		{
			name: "no metric name",
			metric: models.Metrics{
				Type: models.Counter,
			},
			wantErr: true,
		},
		{
			name: "invalid metric type",
			metric: models.Metrics{
				Name: "someMetric",
				Type: "invalidType",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMetric(&tt.metric)
			if tt.wantErr {
				require.Error(t, err, "Error expected")
			} else {
				require.NoError(t, err, "No error expected, got '%v'", err)
			}
		})
	}
}

// for testing ping
type fakePStore struct {
	store.PersistentStorage
	Pingable bool
}

func (ps fakePStore) Ping() error {
	if ps.Pingable {
		return nil
	}
	return errors.New("error")
}

func TestHTTPHandlers_PingHandler(t *testing.T) {
	s := &memory.MemStorage{}
	h := NewHTTPHandlers(s)
	pStore := &fakePStore{
		PersistentStorage: &file.FileStorage{},
	}
	tests := []struct {
		name     string
		pingable bool
		respCode int
		pstore   store.PersistentStorage
	}{
		{name: "ping ok", pingable: true, respCode: http.StatusOK, pstore: pStore},
		{name: "ping", pingable: false, respCode: http.StatusInternalServerError, pstore: pStore},
		{name: "no pstore", pingable: false, respCode: http.StatusInternalServerError, pstore: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pStore.Pingable = tt.pingable
			h.SetPStorage(tt.pstore)
			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
			w := httptest.NewRecorder()
			h.PingHandler(w, req)
			res := w.Result()
			defer res.Body.Close()
			require.Equal(t, tt.respCode, res.StatusCode)
		})
	}

}
