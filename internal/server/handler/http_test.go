package handler

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/freepaddler/yap-metrics/internal/models"
	"github.com/freepaddler/yap-metrics/internal/store/memory"
)

func TestMetricsServer_IndexHandler(t *testing.T) {
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

func TestMetricsServer_UpdateHandler(t *testing.T) {
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

func TestMetricsServer_ValueHandler(t *testing.T) {
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
