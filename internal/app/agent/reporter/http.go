package reporter

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/freepaddler/yap-metrics/internal/pkg/compress"
	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
	"github.com/freepaddler/yap-metrics/internal/pkg/sign"
	"github.com/freepaddler/yap-metrics/internal/pkg/store"
	"github.com/freepaddler/yap-metrics/pkg/retry"
)

// HTTPReporter reports metrics to server over HTTP
type HTTPReporter struct {
	storage store.Storage
	address string
	client  http.Client
	key     string
}

func NewHTTPReporter(s store.Storage, address string, timeout time.Duration, key string) *HTTPReporter {
	return &HTTPReporter{
		storage: s,
		address: address,
		client:  http.Client{Timeout: timeout},
		key:     key,
	}
}

func (r HTTPReporter) ReportBatchJSON(ctx context.Context) {
	logger.Log().Debug().Msg("reporting metrics")
	// get storage snapshot
	m := r.storage.Snapshot(true)
	if len(m) == 0 {
		logger.Log().Info().Msg("nothing to report, skipping")
		return
	}

	err := retry.WithStrategy(ctx,
		func(context.Context) error {
			err := func() (err error) {
				logger.Log().Debug().Msgf("sending %d metrics in batch", len(m))
				url := fmt.Sprintf("http://%s/updates/", r.address)
				body, err := json.Marshal(m)
				if err != nil {
					logger.Log().Warn().Err(err).Msg("unable to marshal JSON batch")
					return
				}

				// calculate hash
				var HashSHA256 string
				if r.key != "" {
					HashSHA256 = sign.Get(body, r.key)
				}

				// compress body
				reqBody, compressErr := compress.GzipBody(&body)

				req, err := http.NewRequest(http.MethodPost, url, reqBody)
				if err != nil {
					logger.Log().Error().Err(err).Msg("unable to create http request")
					return
				}
				// set hash header
				if HashSHA256 != "" {
					req.Header.Set("HashSHA256", HashSHA256)
				}
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Accept-Encoding", "gzip")
				if compressErr == nil {
					req.Header.Set("Content-Encoding", "gzip")
				}
				logger.Log().Debug().Msgf("sending metric %s", body)
				resp, err := r.client.Do(req)
				if err != nil {
					logger.Log().Warn().Err(err).Msgf("failed to send metric %s", body)
					return
				}
				defer resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					// request failed
					logger.Log().Warn().Msgf("wrong http response status: %s", resp.Status)
					return
				}
				return nil
			}()
			return err
		},
		retry.IsNetErr,
		1, 3, 5,
	)
	if err != nil {
		logger.Log().Warn().Err(err).Msg("report failed")
	}
}
