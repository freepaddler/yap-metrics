package sign

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"net/http"

	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
)

// Get returns base64 encoded HMAC SHA-256 hash
func Get(data []byte, key string) string {
	hash := hmac.New(sha256.New, []byte(key))
	hash.Write(data)
	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}

// RespWrapper replaces http ResponseWriter Write method
type RespWrapper struct {
	http.ResponseWriter
	key string
}

func NewRespWrapper(w http.ResponseWriter, k string) *RespWrapper {
	return &RespWrapper{
		ResponseWriter: w,
		key:            k,
	}
}

// Write calculates and sets HashSHA256 header
func (rw RespWrapper) Write(b []byte) (int, error) {
	HashSHA256 := Get(b, rw.key)
	rw.Header().Add("HashSHA256", HashSHA256)
	logger.Log.Debug().Msg("hash calculated, header HashSHA256 added")
	return rw.ResponseWriter.Write(b)
}

// TODO: test

// Middleware to check signature of request and add signature of response
func Middleware(key string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		sigmw := func(w http.ResponseWriter, r *http.Request) {

			// wrapped response writer to proceed with
			ww := w

			// proceed request if key is set and exists in header
			reqSign := r.Header.Get("HashSHA256")
			if key != "" && reqSign != "" {
				logger.Log.Debug().Msgf("sign: key '%s', HashSHA256 '%s'", key, reqSign)
				reqBody, err := io.ReadAll(r.Body)
				if err != nil {
					logger.Log.Warn().Err(err).Msg("failed to read request body")
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				defer r.Body.Close()
				bodySign := Get(reqBody, key)
				logger.Log.Debug().Msgf("signed body is '%s'", bodySign)
				if reqSign != bodySign {
					logger.Log.Warn().Err(err).Msg("invalid HashSHA256 signature")
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				// return body to be read from handler
				r.Body = io.NopCloser(bytes.NewBuffer(reqBody))
				// key is set, we need to work with response
				ww = NewRespWrapper(w, key)
				logger.Log.Debug().Msg("signature validated")
			}

			next.ServeHTTP(ww, r)

		}
		return http.HandlerFunc(sigmw)
	}
}
