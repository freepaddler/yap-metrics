package reporter

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
	"github.com/freepaddler/yap-metrics/internal/pkg/models"
	"github.com/freepaddler/yap-metrics/internal/pkg/store"
)

// HTTPReporter reports metrics to server over HTTP
type HTTPReporter struct {
	storage store.Storage
	address string
	client  http.Client
}

func NewHTTPReporter(s store.Storage, address string, timeout time.Duration) *HTTPReporter {
	return &HTTPReporter{
		storage: s,
		address: address,
		client:  http.Client{Timeout: timeout},
	}
}

func (r HTTPReporter) ReportJSON() {
	logger.Log.Debug().Msg("ReportJSON: reporting metrics")
	// get storage snapshot
	m := r.storage.Snapshot(true)

	url := fmt.Sprintf("http://%s/update", r.address)
	for _, v := range m {
		// returns false if metric was not successfully reported to server
		reported := func() bool {
			body, err := json.Marshal(v)
			if err != nil {
				logger.Log.Warn().Err(err).Msgf("unable to marshal JSON: %+v", v)
				return false
			}

			// compress body
			respBody, compressErr := compressResponse(&body)

			req, err := http.NewRequest(http.MethodPost, url, respBody)
			if err != nil {
				logger.Log.Error().Err(err).Msg("unable to create http request")
				return false
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept-Encoding", "gzip")
			if compressErr == nil {
				req.Header.Set("Content-Encoding", "gzip")
			}
			logger.Log.Debug().Msgf("sending metric %s", body)
			resp, err := r.client.Do(req)
			if err != nil {
				logger.Log.Warn().Err(err).Msgf("failed to send metric %s", body)
				return false
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				// request failed
				logger.Log.Warn().Msgf("wrong http response status: %s", resp.Status)
				return false
			}
			return true
		}()
		// restore unsent metric back to storage
		if !reported {
			logger.Log.Debug().Msgf("restore metric %+v back to storage", v)
			r.storage.RestoreMetrics([]models.Metrics{v})
		}
	}
}

func (r HTTPReporter) ReportBatchJSON() {
	logger.Log.Debug().Msg("reporting metrics")
	// get storage snapshot
	m := r.storage.Snapshot(true)

	reported := func() bool {
		if len(m) == 0 {
			logger.Log.Info().Msg("nothing to report, skipping")
			return false
		}
		logger.Log.Debug().Msgf("sending %d metrics in batch", len(m))
		url := fmt.Sprintf("http://%s/updates/", r.address)
		body, err := json.Marshal(m)
		if err != nil {
			logger.Log.Warn().Err(err).Msg("unable to marshal JSON batch")
			return false
		}
		// compress body
		respBody, compressErr := compressResponse(&body)

		req, err := http.NewRequest(http.MethodPost, url, respBody)
		if err != nil {
			logger.Log.Error().Err(err).Msg("unable to create http request")
			return false
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept-Encoding", "gzip")
		if compressErr == nil {
			req.Header.Set("Content-Encoding", "gzip")
		}
		logger.Log.Debug().Msgf("sending metric %s", body)
		resp, err := r.client.Do(req)
		if err != nil {
			logger.Log.Warn().Err(err).Msgf("failed to send metric %s", body)
			return false
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			// request failed
			logger.Log.Warn().Msgf("wrong http response status: %s", resp.Status)
			return false
		}
		return true
	}()

	if !reported {
		logger.Log.Debug().Msgf("restore %d metrics back to storage", len(m))
		r.storage.RestoreMetrics(m)
	}

}

func compressResponse(body *[]byte) (*bytes.Buffer, error) {
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
