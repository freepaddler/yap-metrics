package httpbatchreporter

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/freepaddler/yap-metrics/internal/pkg/models"
)

func TestHTTPReporter_ReportBatchJSON(t *testing.T) {

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
		},
		{
			name:            "success report",
			httpRequest:     true,
			serverReachable: true,
			respCode:        http.StatusOK,
		},
		{
			name:            "signed report",
			httpRequest:     true,
			serverReachable: true,
			key:             "someKey",
			respCode:        http.StatusOK,
		},
		{
			name:            "server bad response",
			httpRequest:     true,
			serverReachable: true,
			respCode:        http.StatusBadRequest,
		},
		{
			name:            "server unreachable",
			serverReachable: false,
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
			h := New(
				WithAddress(address),
				WithHTTPTimeout(time.Second),
				WithSignKey(tt.key),
			)

			h.Send(report)
		})
	}
}
