package ipmatcher

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIPMatchMiddleware(t *testing.T) {
	tests := []struct {
		name    string
		header  string
		enabled bool
		subnet  netip.Prefix
		want    int
	}{
		{
			name: "check disabled",
			want: http.StatusOK,
		},
		{
			name:    "match ip",
			header:  "10.10.10.10",
			enabled: true,
			subnet:  netip.MustParsePrefix("10.0.0.0/8"),
			want:    http.StatusOK,
		},
		{
			name:    "enabled no header",
			enabled: true,
			subnet:  netip.MustParsePrefix("0.0.0.0/0"),
			want:    http.StatusForbidden,
		},
		{
			name:    "ip doesn't match",
			header:  "10.10.10.10",
			enabled: true,
			subnet:  netip.MustParsePrefix("192.168.0.0/16"),
			want:    http.StatusForbidden,
		},
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				req.Header.Set("x-real-ip", tt.header)
			}
			w := httptest.NewRecorder()
			mw := IPMatchMiddleware(tt.enabled, tt.subnet)
			handler := mw(testHandler)
			handler.ServeHTTP(w, req)

			require.Equal(t, tt.want, w.Code, "code expected %d, got %d", tt.want, w.Code)
		})
	}
}
