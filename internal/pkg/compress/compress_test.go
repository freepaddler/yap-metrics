package compress

import (
	"bytes"
	gzip "compress/gzip"
	"crypto/rand"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGzipBodySuccess(t *testing.T) {
	body := make([]byte, 1024)
	// create random body
	rand.Read(body)
	// duplicate it to be sure in compression
	body = append(body, body...)
	got, err := GzipBody(&body)
	require.NoError(t, err)
	assert.Less(t, got.Len(), len(body), "Expected result size %d less than source size %d", got.Len(), len(body))
	gunzipped, err := gzip.NewReader(got)
	if err != nil {
		require.NoError(t, err, "Got error '%v' on decompression", err)
	}
	defer gunzipped.Close()
	body2, err := io.ReadAll(gunzipped)
	require.NoError(t, err, "Got error '%v' reading decompressed body", err)
	assert.Equal(t, body, body2, "Expected equal, got different")
}

func TestGunzipMiddleware(t *testing.T) {
	body := []byte("this is a test body")
	body = append(body, body...)
	body = append(body, body...)

	plainBody := bytes.NewReader(body)
	gzipBody, gzErr := GzipBody(&body)
	assert.NoError(t, gzErr, "Expected no error on compression, got '%v'", gzErr)

	tests := []struct {
		name    string
		body    io.Reader
		header  bool
		resCode int
	}{
		{
			name:    "uncompressed body",
			body:    plainBody,
			resCode: http.StatusOK,
		},
		{
			name:    "compressed body",
			body:    gzipBody,
			resCode: http.StatusOK,
			header:  true,
		},
		{
			name:    "bad compressed body",
			body:    plainBody,
			resCode: http.StatusInternalServerError,
			header:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var resBody []byte
			var err error
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				resBody, err = io.ReadAll(r.Body)
				defer r.Body.Close()
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
				}
				w.WriteHeader(http.StatusOK)
			})
			req := httptest.NewRequest(http.MethodPost, "/", tt.body)
			if tt.header {
				req.Header.Add("Content-Encoding", "gzip")
			}
			w := httptest.NewRecorder()
			handler := GunzipMiddleware(testHandler)
			handler.ServeHTTP(w, req)
			res := w.Result()
			defer res.Body.Close()
			require.Equal(t, tt.resCode, res.StatusCode, "Expected result %d, got %d", tt.resCode, res.StatusCode)
			if res.StatusCode == http.StatusOK {
				assert.Equal(t, body, resBody, "Expected request and response body to be equal")
			}
		})
	}
	//var resBody []byte
	//var err error
	//testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//	resBody, err = io.ReadAll(r.Body)
	//	defer r.Body.Close()
	//	if err != nil {
	//		w.WriteHeader(http.StatusBadRequest)
	//	}
	//	w.WriteHeader(http.StatusOK)
	//})
	//body := []byte("this is a test body")
	//body = append(body, body...)
	//body = append(body, body...)
	//reqBody := bytes.NewReader(body)
	//req := httptest.NewRequest(http.MethodPost, "/", reqBody)
	////req.Header.Add("Content-Encoding", "gzip")
	//w := httptest.NewRecorder()
	//handler := GunzipMiddleware(testHandler)
	//handler.ServeHTTP(w, req)
	//res := w.Result()
	//require.Equal(t, http.StatusOK, res.StatusCode, "Expected")
	//require.NoError(t, err)
	//assert.Equal(t, body, resBody)
}
