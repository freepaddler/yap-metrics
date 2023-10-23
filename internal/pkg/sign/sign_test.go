package sign

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSign(t *testing.T) {
	// server response body
	srvBody := []byte("this is a response body")
	// client request body
	clientBody := []byte("this is a request body")
	reqBody := bytes.NewReader(clientBody)

	// sign keys
	key1 := "key1"
	key2 := "key2"

	testHandler := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Write(srvBody)
	})

	tests := []struct {
		name      string
		resCode   int
		clientKey string
		serverKey string
	}{
		{
			name:      "both signed same key",
			resCode:   http.StatusOK,
			clientKey: key1,
			serverKey: key1,
		},
		{
			name:      "request is not signed, server with key",
			resCode:   http.StatusOK,
			serverKey: key1,
		},
		{
			name:      "request signed, no server key",
			resCode:   http.StatusOK,
			clientKey: key1,
		},
		{
			name:    "no keys on client and server",
			resCode: http.StatusOK,
		},
		{
			name:      "different keys",
			resCode:   http.StatusBadRequest,
			clientKey: key1,
			serverKey: key2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// generate request
			req := httptest.NewRequest(http.MethodPost, "/", reqBody)
			var HashSHA256 string
			if tt.clientKey != "" {
				HashSHA256 = Get(clientBody, tt.clientKey)
			}
			if HashSHA256 != "" {
				req.Header.Set("HashSHA256", HashSHA256)
			}

			// middlware
			w := httptest.NewRecorder()
			mw := Middleware(tt.serverKey)
			handler := mw(testHandler)
			handler.ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()

			// check code
			require.Equal(t, tt.resCode, res.StatusCode, "Expected result %d, got %d", http.StatusOK, res.StatusCode)
			if tt.resCode == http.StatusOK {
				// check result signature
				if tt.clientKey != "" && tt.serverKey != "" {
					assert.Equal(
						t,
						Get(srvBody, tt.serverKey),
						res.Header.Get("HashSHA256"),
						"Expected equal signature of server response",
					)
				} else {
					assert.Equal(
						t,
						"",
						res.Header.Get("HashSHA256"),
						"Expected no signature in response if no server key or client request was without signature",
					)
				}

			}
		})
	}
}
