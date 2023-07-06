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

func (r HTTPReporter) Report() {
	m := r.storage.Snapshot()
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
			logger.Log.Debug().Msgf("sending metric %s", url)
			resp, err := r.client.Post(url, "text/plain", nil)
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
				logger.Log.Debug().Msgf("response body: %s", body)
				return
			}
			// request successes, delete updated metrics
			switch v.Type {
			case models.Counter:
				r.storage.DelCounter(v.Name)
			case models.Gauge:
				r.storage.DelGauge(v.Name)

			}
		}()
	}
}

func (r HTTPReporter) ReportJSON() {
	// TODO: add mutex where? snapsot? or snashot + flush?
	m := r.storage.Snapshot()
	r.storage.Flush()

	// TODO: question
	// переименование метода GetAllMetrics в Snapshot показало сломанную логику реализации агента
	// мы отправляем потенциально одно значение (из снапшота), а удаляем другое -
	// из хранилища, которое могло обновиться - та же проблема, что и запись в файл на сервере
	// логично сделать так:
	// при снятии снапшота - удалять метрики, попавшие в снапшот
	// при неуспешной отравке:
	// для gauge: восстанавливать из снапшота, если не появилось нового значения
	// для counter: прибавлять к метрике в хранилище значение из снапшота
	// но(!) в этом случае методы Snapshot для агента и сервера ДОЛЖНЫ быть разными

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
			if updated, _ := r.storage.GetMetric(&v); updated {
				r.storage.SetMetric(&v)
			}
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
