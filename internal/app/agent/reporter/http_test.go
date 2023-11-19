package reporter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/freepaddler/yap-metrics/internal/pkg/models"
	"github.com/freepaddler/yap-metrics/mocks"
)

func TestHTTPReporter_ReportBatchJSON(t *testing.T) {
	var mockController = gomock.NewController(t)
	defer mockController.Finish()
	m := mocks.NewMockReporter(mockController)

	report := []models.Metrics{
		{
			Name:   "c1",
			Type:   models.Counter,
			IValue: new(int64),
		},
		{
			Name:   "g1",
			Type:   models.Gauge,
			FValue: new(float64),
		},
	}

	tests := []struct {
		name            string
		httpRequest     bool
		serverReachable bool
		respCode        int
		key             string
		mocks           func()
	}{
		{
			name:        "empty report",
			httpRequest: false,
			mocks: func() {
				m.EXPECT().ReportAll().Return([]models.Metrics{}, time.Now()).Times(1)
			},
		},
		{
			name:            "success report",
			httpRequest:     true,
			serverReachable: true,
			respCode:        http.StatusOK,
			mocks: func() {
				m.EXPECT().ReportAll().Return(report, time.Now()).Times(1)
			},
		},
		{
			name:            "signed report",
			httpRequest:     true,
			serverReachable: true,
			key:             "someKey",
			respCode:        http.StatusOK,
			mocks: func() {
				m.EXPECT().ReportAll().Return(report, time.Now()).Times(1)
			},
		},
		{
			name:            "server bad response",
			httpRequest:     true,
			serverReachable: true,
			respCode:        http.StatusBadRequest,
			mocks: func() {
				m.EXPECT().ReportAll().Return(report, time.Now()).Times(1)
				m.EXPECT().RestoreReport(report, gomock.AssignableToTypeOf(time.Time{})).Times(1)
			},
		},
		{
			name:            "server unreachable",
			serverReachable: false,
			mocks: func() {
				m.EXPECT().ReportAll().Return(report, time.Now()).Times(1)
				m.EXPECT().RestoreReport(report, gomock.AssignableToTypeOf(time.Time{})).Times(1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				// check if http server should be called
				require.True(t, tt.httpRequest, "No http request should be made")
				assert.Equal(t, "POST", req.Method, "POST method expected, got '%s'", req.Method)
				assert.Equal(t, "/updates/", req.URL.Path, "Expected path '/updates/', got '%s'", req.URL.Path)
				assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
				assert.Equal(t, "gzip", req.Header.Get("Accept-Encoding"))
				assert.Equal(t, "gzip", req.Header.Get("Content-Encoding"))
				if tt.key != "" {
					require.NotEmpty(t, req.Header.Get("HashSHA256"), "Expected sign header 'HashSHA256'")
				}
				rw.WriteHeader(tt.respCode)
			}))
			// Close the server when test finishes
			defer server.Close()
			address := "240.0.0.0:65535"
			if tt.serverReachable {
				serverURL, err := url.Parse(server.URL)
				require.NoError(t, err, "Failed to parse test httpserver address")
				address = serverURL.Host
			}

			h := NewHTTPReporter(m, address, time.Second, tt.key)
			tt.mocks()
			h.ReportBatchJSON(context.Background())
			//require.NotPanics(t, func() { h.ReportBatchJSON(context.Background()) })
		})
	}
}
