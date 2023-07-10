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
			// if gauge was already updated, don't update it
			if updated, _ := r.storage.GetMetric(&v); v.Type == models.Gauge && updated {
				break
			}
			r.storage.UpdateMetrics([]models.Metrics{v}, false)
		}
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
