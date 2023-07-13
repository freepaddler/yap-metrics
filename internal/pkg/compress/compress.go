package compress

import (
	"bytes"
	"compress/gzip"
	"net/http"

	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
)

// GunzipMiddleware unzip incoming request
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

func CompressBody(body *[]byte) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	gzBuf, _ := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
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
