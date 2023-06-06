package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/freepaddler/yap-metrics/internal/models"
	"github.com/freepaddler/yap-metrics/internal/store"
)

func TestMetricsServer_IndexHandler(t *testing.T) {
	srv := &MetricsServer{
		storage: store.NewMemStorage(),
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	srv.IndexHandler(w, req)
	res := w.Result()

	require.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "text/html; charset=utf-8", res.Header.Get("Content-Type"))
}

func TestMetricsServer_UpdateHandler(t *testing.T) {
	srv := &MetricsServer{
		storage: store.NewMemStorage(),
	}
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
			mValue: "-0.00117",
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
			srv.UpdateHandler(w, req)
			res := w.Result()

			assert.Equal(t, tt.code, res.StatusCode)
		})
	}
}

func TestMetricsServer_ValueHandler(t *testing.T) {
	srv := &MetricsServer{
		storage: store.NewMemStorage(),
	}
	var cValue int64 = 10
	var cName = "c1"
	var gValue float64 = -0.0017
	var gName = "g1"
	srv.storage.CounterSet(cName, cValue)
	srv.storage.GaugeSet(gName, gValue)
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
			want:  fmt.Sprintf("%d", cValue),
		},
		{
			name:  "success gauge",
			code:  http.StatusOK,
			mType: models.Gauge,
			mName: gName,
			want:  fmt.Sprintf("%f", gValue),
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
			srv.ValueHandler(w, req)
			res := w.Result()

			require.Equal(t, tt.code, res.StatusCode)
			if res.StatusCode == http.StatusOK {
				defer res.Body.Close()
				resBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				assert.Equal(t, tt.want, string(resBody))
			}
		})
	}
}
