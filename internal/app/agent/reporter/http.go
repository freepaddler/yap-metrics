package reporter

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
	"github.com/freepaddler/yap-metrics/internal/pkg/store"
	"github.com/freepaddler/yap-metrics/internal/pkg/store/retry"
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

func (r HTTPReporter) ReportBatchJSON(ctx context.Context) {
	logger.Log.Debug().Msg("reporting metrics")
	// get storage snapshot
	m := r.storage.Snapshot(true)
	if len(m) == 0 {
		logger.Log.Info().Msg("nothing to report, skipping")
		return
	}

	var reported bool
	err := retry.WithStrategy(ctx,
		func(context.Context) error {
			err := func(*bool) (err error) {
				logger.Log.Debug().Msgf("sending %d metrics in batch", len(m))
				url := fmt.Sprintf("http://%s/updates/", r.address)
				body, err := json.Marshal(m)
				if err != nil {
					logger.Log.Warn().Err(err).Msg("unable to marshal JSON batch")
					return
				}
				// compress body
				respBody, compressErr := compressResponse(&body)

				req, err := http.NewRequest(http.MethodPost, url, respBody)
				if err != nil {
					logger.Log.Error().Err(err).Msg("unable to create http request")
					return
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
					return
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					// request failed
					logger.Log.Warn().Msgf("wrong http response status: %s", resp.Status)
					return
				}
				reported = true
				return nil
			}(&reported)
			return err
		},
		retry.IsNetErr,
		1, 3, 5,
	)
	if err != nil {
		logger.Log.Warn().Err(err).Msg("report failed")
	}

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
