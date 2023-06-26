package compress

import (
	"compress/gzip"
	"net/http"
	"strings"

	"github.com/freepaddler/yap-metrics/internal/logger"
)

type gzWriter struct {
	http.ResponseWriter
	gzw *gzip.Writer
}

func newGzWriter(w http.ResponseWriter) *gzWriter {
	gz, _ := gzip.NewWriterLevel(w, gzip.BestSpeed)
	return &gzWriter{
		ResponseWriter: w,
		gzw:            gz,
	}
}

func (cw *gzWriter) Write(p []byte) (n int, err error) {
	ct := cw.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") && !strings.Contains(ct, "application/json") {
		logger.Log.Debug().Msg("sending compressed response")
		return cw.ResponseWriter.Write(p)
	}
	cw.Header().Set("Content-Encoding", "gzip")
	return cw.gzw.Write(p)
}

func (cw *gzWriter) Close() {
	cw.gzw.Close()
}

func GzipMiddleware(next http.Handler) http.Handler {
	gz := func(w http.ResponseWriter, r *http.Request) {
		// compress response: replace ResponseWriter with gzip Writer
		// default response writer
		rw := w

		// client should accept gzip encoding
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			// do not assign to rw directly!!
			gzrw := newGzWriter(w)
			rw = gzrw
			// because it should be closed
			defer gzrw.Close()
		}

		// proceed compressed request
		if r.Header.Get("Content-Encoding") == "gzip" {
			zrBody, err := gzip.NewReader(r.Body)
			if err != nil {
				logger.Log.Warn().Err(err).Msg("failed to create gzip reader")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			defer zrBody.Close()
			logger.Log.Debug().Msgf("gzip-compressed request")
			r.Body = zrBody
		}

		next.ServeHTTP(rw, r)

	}
	return http.HandlerFunc(gz)
}
