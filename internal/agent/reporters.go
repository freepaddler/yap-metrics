package agent

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/freepaddler/yap-metrics/internal/models"
)

type HTTPReporter struct {
	address string
	c       http.Client
}

func NewHTTPReporter(address string, timeout time.Duration) *HTTPReporter {
	return &HTTPReporter{
		address: address,
		c:       http.Client{Timeout: timeout * time.Second},
	}
}

func (r *HTTPReporter) Report(m models.Metrics) bool {
	var val string
	switch m.Type {
	case models.Gauge:
		val = strconv.FormatFloat(m.Gauge, 'f', -1, 64)
	case models.Counter:
		val = strconv.FormatInt(m.Increment, 10)
	}
	url := fmt.Sprintf("http://%s/update/%s/%s/%s", r.address, m.Type, m.Name, val)
	fmt.Printf("Sending metric %s\n", url)
	resp, err := r.c.Post(url, "text/plain", nil)
	if err != nil {
		fmt.Printf("Failed to send metric with error: %s\n", err)
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Got http status: %s\n", resp.Status)

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Unable to parse response body with error: %s\n", err)
		}
		fmt.Printf("Response body: %s\n", body)
		return false
	}
	return true
}

// PrintReporter is a test reporter to stdout
type PrintReporter struct {
}

func NewPrintReporter() *PrintReporter {
	return &PrintReporter{}
}

func (r PrintReporter) Report(m models.Metrics) bool {
	switch m.Type {
	case models.Counter:
		fmt.Printf("Metric: %s, type: %s, value: %d\n", m.Name, m.Type, m.Increment)
	case models.Gauge:
		fmt.Printf("Metric: %s, type: %s, value: %f\n", m.Name, m.Type, m.Gauge)
	}
	return true
}
