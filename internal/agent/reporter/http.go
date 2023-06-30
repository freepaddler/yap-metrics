package reporter

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/freepaddler/yap-metrics/internal/logger"
	"github.com/freepaddler/yap-metrics/internal/models"
	"github.com/freepaddler/yap-metrics/internal/store"
)

// HTTPReporter reports metrics to server over HTTP
type HTTPReporter struct {
	s       store.Storage
	address string
	c       http.Client
}

func NewHTTPReporter(s store.Storage, address string, timeout time.Duration) *HTTPReporter {
	return &HTTPReporter{
		s:       s,
		address: address,
		c:       http.Client{Timeout: timeout},
	}
}

func (r HTTPReporter) Report() {
	m := r.s.GetAllMetrics()
	for _, v := range m {
		func() {
			var val string
			switch v.Type {
			case models.Gauge:
				val = strconv.FormatFloat(*v.FValue, 'f', -1, 64)
			case models.Counter:
				val = strconv.FormatInt(*v.IValue, 10)
			}
			url := fmt.Sprintf("http://%s/update/%s/%s/%s", r.address, v.Type, v.Name, val)
			logger.Log.Debug().Msgf("Sending metric %s", url)
			resp, err := r.c.Post(url, "text/plain", nil)
			if err != nil {
				logger.Log.Warn().Err(err).Msgf("failed to send metric %s", url)
				return
			}
			defer resp.Body.Close()
			// check if request was successful
			if resp.StatusCode != http.StatusOK {
				// request failed
				logger.Log.Warn().Msgf("wrong http response status: %s", resp.Status)

				body, err := io.ReadAll(resp.Body)
				if err != nil {
					logger.Log.Warn().Err(err).Msg("unable to parse response body")
				}
				logger.Log.Debug().Msgf("Response body: %s", body)
				return
			}
			// request successes, delete updated metrics
			switch v.Type {
			case models.Counter:
				r.s.DelCounter(v.Name)
			case models.Gauge:
				r.s.DelGauge(v.Name)

			}
		}()
	}
}

func (r HTTPReporter) ReportJSON() {
	m := r.s.GetAllMetrics()
	url := fmt.Sprintf("http://%s/update", r.address)
	for _, v := range m {
		func() {
			body, err := json.Marshal(v)
			if err != nil {
				logger.Log.Warn().Err(err).Msgf("unable to marshal JSON: %+v", v)
				return
			}

			// compress body
			respBody := compressResponse(&body)

			req, err := http.NewRequest(http.MethodPost, url, respBody)
			if err != nil {
				logger.Log.Error().Err(err).Msg("unable to create http request")
			}
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Content-Encoding", "gzip")
			req.Header.Set("Accept-Encoding", "gzip")
			logger.Log.Debug().Msgf("sending metric %s", body)
			resp, err := r.c.Do(req)
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
			// request successes, delete updated metrics
			switch v.Type {
			case models.Counter:
				r.s.DelCounter(v.Name)
			case models.Gauge:
				r.s.DelGauge(v.Name)

			}
		}()
	}
}

func compressResponse(body *[]byte) *bytes.Buffer {
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
	return &buf

}
