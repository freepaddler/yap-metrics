package handler

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/freepaddler/yap-metrics/internal/pkg/models"
	"github.com/freepaddler/yap-metrics/internal/pkg/store"
	"github.com/freepaddler/yap-metrics/mocks"
	"github.com/freepaddler/yap-metrics/test/utils"
)

func TestHTTPHandlers_PingHandler(t *testing.T) {
	var mockController = gomock.NewController(t)
	defer mockController.Finish()
	m := mocks.NewMockHTTPHandlerStorage(mockController)

	h := NewHTTPHandlers(m)

	tests := []struct {
		name      string
		wantCode  int
		returnErr error
	}{
		{
			name:     "success ping",
			wantCode: http.StatusOK,
		},
		{
			name:      "ping failed",
			wantCode:  http.StatusInternalServerError,
			returnErr: errors.New("some"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
			m.EXPECT().Ping().Times(1).Return(tt.returnErr)
			w := httptest.NewRecorder()
			h.PingHandler(w, req)
			res := w.Result()
			defer res.Body.Close()
			require.Equal(t, tt.wantCode, res.StatusCode)

		})
	}
}

func TestHTTPHandlers_Index(t *testing.T) {
	var mockController = gomock.NewController(t)
	defer mockController.Finish()
	m := mocks.NewMockHTTPHandlerStorage(mockController)

	h := NewHTTPHandlers(m)

	m.EXPECT().GetAll().Times(1)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.IndexMetricHandler(w, req)
	res := w.Result()
	defer res.Body.Close()

	require.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "text/html; charset=utf-8", res.Header.Get("Content-Type"))
}

func TestHTTPHandlers_GetMetric(t *testing.T) {
	var mockController = gomock.NewController(t)
	defer mockController.Finish()
	m := mocks.NewMockHTTPHandlerStorage(mockController)

	h := NewHTTPHandlers(m)

	tests := []struct {
		name string

		mName       string
		mType       string
		counterVal  int64   // return int value
		gaugeVal    float64 // return float value
		wantValue   string
		wantCode    int
		wantCall    int
		returnError error
	}{
		{
			name:       "success counter",
			mName:      "name",
			mType:      "counter",
			counterVal: 10,

			wantValue: "10",
			wantCode:  http.StatusOK,
			wantCall:  1,
		},
		{
			name:     "success gauge",
			mName:    "name",
			mType:    "gauge",
			gaugeVal: 0.119,

			wantValue: "0.119",
			wantCode:  http.StatusOK,
			wantCall:  1,
		},
		{
			name:     "invalid type",
			wantCode: http.StatusBadRequest,
			mType:    "qqq",
			mName:    "name",
			wantCall: 0,
		},
		{
			name:        "not found",
			wantCode:    http.StatusNotFound,
			mType:       "counter",
			mName:       "some name",
			wantCall:    1,
			returnError: store.ErrMetricNotFound,
		},
		{
			name:     "no name",
			wantCode: http.StatusBadRequest,
			mType:    "counter",
			wantCall: 0,
		},
		{
			name:     "no type",
			wantCode: http.StatusBadRequest,
			mName:    "models.Counter",
			wantCall: 0,
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
			m.EXPECT().GetOne(models.MetricRequest{
				Name: tt.mName,
				Type: tt.mType,
			}).Times(tt.wantCall).DoAndReturn(func(req models.MetricRequest) (m models.Metrics, err error) {
				return models.Metrics{
					Type:   tt.mType,
					FValue: &tt.gaugeVal,
					IValue: &tt.counterVal,
				}, tt.returnError
			})
			h.GetMetricHandler(w, req)
			res := w.Result()
			defer res.Body.Close()

			require.Equal(t, tt.wantCode, res.StatusCode)
			if res.StatusCode == http.StatusOK {
				resBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				assert.Equal(t, tt.wantValue, string(resBody))
			}
		})
	}
}

func TestHTTPHandlers_GetMetricJSON(t *testing.T) {
	var mockController = gomock.NewController(t)
	defer mockController.Finish()
	m := mocks.NewMockHTTPHandlerStorage(mockController)

	h := NewHTTPHandlers(m)

	tests := []struct {
		name string

		rawRequest  string               // http request body
		wantRequest models.MetricRequest // mock request

		counterVal  *int64   // return int value
		gaugeVal    *float64 // return float value
		want        string
		wantCode    int
		wantCall    int
		returnError error
	}{
		{
			name:        "success counter",
			rawRequest:  `{"id":"name","type":"counter"}`,
			wantRequest: models.MetricRequest{Name: "name", Type: "counter"},
			counterVal:  utils.Pointer(int64(10)),

			want:     `{"id":"name","type":"counter","delta":10}`,
			wantCode: http.StatusOK,
			wantCall: 1,
		},
		{
			name:        "success gauge",
			rawRequest:  `{"id":"name","type":"gauge"}`,
			wantRequest: models.MetricRequest{Name: "name", Type: "gauge"},
			gaugeVal:    utils.Pointer(-1000.0001),

			want:     `{"id":"name","type":"gauge","value":-1000.0001}`,
			wantCode: http.StatusOK,
			wantCall: 1,
		},
		{
			name:       "invalid type",
			wantCode:   http.StatusBadRequest,
			rawRequest: `{"id":"name","type":"gauge1"}`,
			wantCall:   0,
		},
		{
			name:        "not found",
			wantCode:    http.StatusNotFound,
			rawRequest:  `{"id":"name","type":"gauge"}`,
			wantRequest: models.MetricRequest{Name: "name", Type: "gauge"},
			wantCall:    1,
			returnError: store.ErrMetricNotFound,
		},
		{
			name:       "no name",
			wantCode:   http.StatusBadRequest,
			rawRequest: `{"type":"gauge"}`,
			wantCall:   0,
		},
		{
			name:       "no type",
			wantCode:   http.StatusBadRequest,
			rawRequest: `{"id":"name"}`,
			wantCall:   0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := bytes.NewBuffer([]byte(tt.rawRequest))
			req := httptest.NewRequest(http.MethodPost, "/value", reqBody)

			w := httptest.NewRecorder()
			m.EXPECT().GetOne(tt.wantRequest).Times(tt.wantCall).DoAndReturn(func(req models.MetricRequest) (m models.Metrics, err error) {
				return models.Metrics{
					Name:   tt.wantRequest.Name,
					Type:   tt.wantRequest.Type,
					FValue: tt.gaugeVal,
					IValue: tt.counterVal,
				}, tt.returnError
			})
			h.GetMetricJSONHandler(w, req)
			res := w.Result()
			defer res.Body.Close()

			require.Equal(t, tt.wantCode, res.StatusCode)
			if res.StatusCode == http.StatusOK {
				resBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				assert.JSONEq(t, tt.want, string(resBody))
			}
		})
	}

}

func TestHTTPHandlers_UpdateMetric(t *testing.T) {
	var mockController = gomock.NewController(t)
	defer mockController.Finish()
	m := mocks.NewMockHTTPHandlerStorage(mockController)

	h := NewHTTPHandlers(m)

	tests := []struct {
		name string

		mName  string
		mType  string
		mValue string   // value in request
		delta  *int64   // delta in mock request
		value  *float64 // value in mock request

		counterVal int64   // return int value
		gaugeVal   float64 // return float value

		wantValue   string
		wantCode    int
		wantCall    int
		returnError error
	}{
		{
			name:   "success counter",
			mName:  "name",
			mType:  "counter",
			mValue: "12",
			delta:  utils.Pointer(int64(12)),

			counterVal: 10,
			wantValue:  "10",

			wantCode: http.StatusOK,
			wantCall: 1,
		},
		{
			name:   "success gauge",
			mName:  "name",
			mType:  "gauge",
			mValue: "1.117",
			value:  utils.Pointer(1.117),

			gaugeVal:  0.119,
			wantValue: "0.119",
			wantCode:  http.StatusOK,
			wantCall:  1,
		},
		{
			name:     "invalid type",
			wantCode: http.StatusBadRequest,
			mType:    "qqq",
			mName:    "name",
			mValue:   "1.117",
			wantCall: 0,
		},
		{
			name:     "no name",
			wantCode: http.StatusBadRequest,
			mType:    "counter",
			mValue:   "10",
			wantCall: 0,
		},
		{
			name:     "no type",
			wantCode: http.StatusBadRequest,
			mName:    "models.Counter",
			mValue:   "100",
			wantCall: 0,
		},
		{
			name:     "no value",
			wantCode: http.StatusBadRequest,
			mName:    "models.Counter",
			mType:    "gauge",
			wantCall: 0,
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
			m.EXPECT().
				UpdateOne(gomock.Any()).
				Times(tt.wantCall).
				DoAndReturn(func(metric *models.Metrics) error {
					assert.Equal(t, tt.mName, metric.Name)
					assert.Equal(t, tt.mType, metric.Type)
					if tt.delta != nil {
						assert.Equal(t, *tt.delta, *metric.IValue)
						*metric.IValue = tt.counterVal
					}
					if tt.value != nil {
						assert.Equal(t, *tt.value, *metric.FValue)
						*metric.FValue = tt.gaugeVal
					}
					return tt.returnError
				})
			h.UpdateMetricHandler(w, req)
			res := w.Result()
			defer res.Body.Close()

			require.Equal(t, tt.wantCode, res.StatusCode)
			if res.StatusCode == http.StatusOK {
				resBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				assert.Equal(t, tt.wantValue, string(resBody))
			}
		})
	}
}

func TestHTTPHandlers_UpdateMetricJSON(t *testing.T) {
	var mockController = gomock.NewController(t)
	defer mockController.Finish()
	m := mocks.NewMockHTTPHandlerStorage(mockController)

	h := NewHTTPHandlers(m)

	tests := []struct {
		name string

		rawRequest  string         // http request body
		wantRequest models.Metrics // mock request

		counterVal *int64   // return int value
		gaugeVal   *float64 // return float value

		want        string
		wantCode    int
		wantCall    int
		returnError error
	}{
		{
			name:        "success counter",
			rawRequest:  `{"id":"name","type":"counter","delta":10}`,
			wantRequest: models.Metrics{Name: "name", Type: "counter", IValue: utils.Pointer(int64(10))},
			counterVal:  utils.Pointer(int64(12)),

			want:     `{"id":"name","type":"counter","delta":12}`,
			wantCode: http.StatusOK,
			wantCall: 1,
		},
		{
			name:        "success gauge",
			rawRequest:  `{"id":"name","type":"gauge","value":0.197}`,
			wantRequest: models.Metrics{Name: "name", Type: "gauge", FValue: utils.Pointer(0.197)},
			gaugeVal:    utils.Pointer(-19.19),

			want:     `{"id":"name","type":"gauge","value":-19.19}`,
			wantCode: http.StatusOK,
			wantCall: 1,
		},
		{
			name:       "invalid type",
			wantCode:   http.StatusBadRequest,
			rawRequest: `{"id":"name","type":"gauge1","value":0.197}`,
			wantCall:   0,
		},
		{
			name:       "no name",
			wantCode:   http.StatusBadRequest,
			rawRequest: `{"type":"gauge1","value":0.197}`,
			wantCall:   0,
		},
		{
			name:       "no type",
			wantCode:   http.StatusBadRequest,
			rawRequest: `{"id":"name","delta":10}`,
			wantCall:   0,
		},
		{
			name:       "no value",
			wantCode:   http.StatusBadRequest,
			rawRequest: `{"id":"name","type":"gauge"}`,
			wantCall:   0,
		},
		{
			name:       "value and delta",
			wantCode:   http.StatusBadRequest,
			rawRequest: `{"id":"name","type":"gauge1","delta":10,"value":19}`,
			wantCall:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := bytes.NewBuffer([]byte(tt.rawRequest))
			req := httptest.NewRequest(http.MethodPost, "/update", reqBody)

			w := httptest.NewRecorder()
			m.EXPECT().
				UpdateOne(gomock.Any()).
				Times(tt.wantCall).
				DoAndReturn(func(metric *models.Metrics) error {
					assert.Equal(t, tt.wantRequest.Name, metric.Name)
					assert.Equal(t, tt.wantRequest.Type, metric.Type)
					if tt.wantRequest.IValue != nil {
						assert.Equal(t, *tt.wantRequest.IValue, *metric.IValue)
						*metric.IValue = *tt.counterVal
					}
					if tt.wantRequest.FValue != nil {
						assert.Equal(t, *tt.wantRequest.FValue, *metric.FValue)
						*metric.FValue = *tt.gaugeVal
					}
					return tt.returnError
				})
			h.UpdateMetricJSONHandler(w, req)
			res := w.Result()
			defer res.Body.Close()

			require.Equal(t, tt.wantCode, res.StatusCode)
			if res.StatusCode == http.StatusOK {
				resBody, err := io.ReadAll(res.Body)
				require.NoError(t, err)
				assert.JSONEq(t, tt.want, string(resBody))
			}
		})
	}
}

func TestHTTPHandlers_UpdateMetricBatch(t *testing.T) {
	var mockController = gomock.NewController(t)
	defer mockController.Finish()
	m := mocks.NewMockHTTPHandlerStorage(mockController)

	h := NewHTTPHandlers(m)

	tests := []struct {
		name string

		rawRequest  string           // http request body
		wantRequest []models.Metrics // mock request

		wantCode    int
		wantCall    int
		returnError error
	}{
		{
			name: "success",
			rawRequest: `[
				{"id":"name","type":"gauge","value":19.1970},
				{"id":"name","type":"counter","delta":-10}
			]`,
			wantRequest: []models.Metrics{
				{Name: "name", Type: "gauge", FValue: utils.Pointer(19.1970)},
				{Name: "name", Type: "counter", IValue: utils.Pointer(int64(-10))},
			},
			wantCode: http.StatusOK,
			wantCall: 1,
		},
		{
			name:       "empty body",
			rawRequest: ``,
			wantCode:   http.StatusBadRequest,
		},
		{
			name:       "empty array",
			rawRequest: `[]`,
			wantCode:   http.StatusBadRequest,
		},
		{
			name:       "invalid json",
			rawRequest: `{name:value}`,
			wantCode:   http.StatusBadRequest,
		},
		{
			name: "invalid type",
			rawRequest: `[
				{"id":"name","type":"gauge1","value":19.1970},
				{"id":"name","type":"counter","delta":-10}
			]`,
			wantCode: http.StatusBadRequest,
		},
		{
			name: "invalid counter value",
			rawRequest: `[
				{"id":"name","type":"gauge","value":19.1970},
				{"id":"name","type":"counter","delta":-10.1}
			]`,
			wantCode: http.StatusBadRequest,
		},
		{
			name: "invalid gauge value",
			rawRequest: `[
				{"id":"name","type":"gauge","value":"abc"},
				{"id":"name","type":"counter","delta":-10}
			]`,
			wantCode: http.StatusBadRequest,
		},
		{
			name: "missing value",
			rawRequest: `[
				{"id":"name","type":"gauge"},
				{"id":"name","type":"counter","delta":-10}
			]`,
			wantCode: http.StatusBadRequest,
		},
		{
			name: "missing type",
			rawRequest: `[
				{"id":"name","type":"gauge","value":19.1970},
				{"id":"name","delta":-10}
			]`,
			wantCode: http.StatusBadRequest,
		},
		{
			name: "missing name",
			rawRequest: `[
				{"id":"name","type":"gauge","value":19.1970},
				{"type":"counter","delta":-10}
			]`,
			wantCode: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := bytes.NewBuffer([]byte(tt.rawRequest))
			req := httptest.NewRequest(http.MethodPost, "/updates", reqBody)

			w := httptest.NewRecorder()
			m.EXPECT().
				UpdateMany(gomock.Any()).
				Times(tt.wantCall).
				DoAndReturn(func(metrics []models.Metrics) error {
					assert.Equal(t, tt.wantRequest, metrics)
					return tt.returnError
				})
			h.UpdateMetricsBatchHandler(w, req)
			res := w.Result()
			defer res.Body.Close()

			require.Equal(t, tt.wantCode, res.StatusCode)
		})
	}
}

// TODO: ping tests
//// for testing ping
//type fakePStore struct {
//	store.PersistentStorage
//	Pingable bool
//}
//
//func (ps fakePStore) Ping() error {
//	if ps.Pingable {
//		return nil
//	}
//	return errors.New("error")
//}
//
//func TestHTTPHandlers_PingHandler(t *testing.T) {
//	s := &memory.MemStorage{}
//	h := NewHTTPHandlers(s)
//	pStore := &fakePStore{
//		PersistentStorage: &file.FileStorage{},
//	}
//	tests := []struct {
//		name     string
//		pingable bool
//		respCode int
//		pstore   store.PersistentStorage
//	}{
//		{name: "ping ok", pingable: true, respCode: http.StatusOK, pstore: pStore},
//		{name: "ping", pingable: false, respCode: http.StatusInternalServerError, pstore: pStore},
//		{name: "no pstore", pingable: false, respCode: http.StatusInternalServerError, pstore: nil},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			pStore.Pingable = tt.pingable
//			h.SetPStorage(tt.pstore)
//			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
//			w := httptest.NewRecorder()
//			h.PingHandler(w, req)
//			res := w.Result()
//			defer res.Body.Close()
//			require.Equal(t, tt.respCode, res.StatusCode)
//		})
//	}
//
//}
