package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func sampleHandler(w http.ResponseWriter, r *http.Request) {
	ct := r.Header.Get("Content-Type")
	rBody, _ := io.ReadAll(r.Body)

	w.Header().Set("Content-Type", ct)
	w.Write(rBody)
}

func TestGzipMiddleware(t *testing.T) {
	testHandler := GzipMiddleware(http.HandlerFunc(sampleHandler))
	srv := httptest.NewServer(testHandler)
	defer srv.Close()

	var sampleJSONBody = `{ "id": "application", "body": "json" }`

	var buf bytes.Buffer
	zb := gzip.NewWriter(&buf)
	_, err := zb.Write([]byte(sampleJSONBody))
	require.NoError(t, err)
	err = zb.Close()
	require.NoError(t, err)
	r := httptest.NewRequest("POST", srv.URL, &buf)
	r.RequestURI = ""
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Content-Encoding", "gzip")
	r.Header.Set("Accept-Encoding", "gzip")

	resp, err := http.DefaultClient.Do(r)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	defer resp.Body.Close()

	zr, err := gzip.NewReader(resp.Body)
	require.NoError(t, err)

	b, err := io.ReadAll(zr)
	require.NoError(t, err)

	require.JSONEq(t, sampleJSONBody, string(b))
}
