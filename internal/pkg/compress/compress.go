// Package compress implements http compression with gzip
package compress

import (
	"bytes"
	"compress/gzip"
	"net/http"

	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
)

// GunzipMiddleware unzips incoming request body.
//
// Usage
//
//	r := chi.NewRouter()
//	r.Use(compress.GunzipMiddleware)
//	r...
func GunzipMiddleware(next http.Handler) http.Handler {
	gz := func(w http.ResponseWriter, r *http.Request) {

		// proceed compressed request
		if r.Header.Get("Content-Encoding") == "gzip" {
			zrBody, err := gzip.NewReader(r.Body)
			if err != nil {
				logger.Log.Warn().Err(err).Msg("failed to create gzip reader")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer zrBody.Close()
			logger.Log.Debug().Msgf("gzip-compressed request decompressed")
			r.Body = zrBody
		}

		next.ServeHTTP(w, r)

	}
	return http.HandlerFunc(gz)
}

// GzipBody tries to gzip response body.
// Returns uncompressed body if compression failed
func GzipBody(body *[]byte) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	// oops, but this is the only thing I could do for a kind of heap optimization :)
	gzBuf, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	defer gzBuf.Close()
	_, err := gzBuf.Write(*body)
	if err != nil {
		logger.Log.Error().Err(err).Msg("unable to compress body, sending uncompressed")
		// return raw body
		buf.Truncate(0)
		buf.Write(*body)
	}
	logger.Log.Debug().Msg("response compressed")
	return &buf, err
}
