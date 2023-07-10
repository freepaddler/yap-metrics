package compress

import (
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
