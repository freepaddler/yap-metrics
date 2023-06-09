package reporter

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

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
		c:       http.Client{Timeout: timeout * time.Second},
	}
}

func (r HTTPReporter) Report() {
	m := r.s.GetAllMetrics()
	for _, v := range m {
		var val string
		switch v.Type {
		case models.Gauge:
			val = strconv.FormatFloat(v.FValue, 'f', -1, 64)
		case models.Counter:
			val = strconv.FormatInt(v.IValue, 10)
		}
		url := fmt.Sprintf("http://%s/update/%s/%s/%s", r.address, v.Type, v.Name, val)
		fmt.Printf("Sending metric %s\n", url)
		resp, err := r.c.Post(url, "text/plain", nil)
		if err != nil {
			fmt.Printf("Failed to send metric with error: %s\n", err)
			continue
		}
		defer resp.Body.Close()
		// check if request was successful
		if resp.StatusCode != http.StatusOK {
			// request failed
			fmt.Printf("Got http status: %s\n", resp.Status)

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("Unable to parse response body with error: %s\n", err)
			}
			fmt.Printf("Response body: %s\n", body)
			continue
		}
		// request successes, delete updated metrics
		switch v.Type {
		case models.Counter:
			r.s.DelCounter(v.Name)
		case models.Gauge:
			r.s.DelGauge(v.Name)

		}
	}
}
