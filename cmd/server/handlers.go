package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/freepaddler/yap-metrics/internal/models"
)

// ValidateMetricURL validates Metrics structure as /method/type/name/value
func ValidateMetricURL(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("ValidateMetricURL: Method:%s, URL:%s\n", r.Method, r.URL)
		// vars[0]="", vars[1]="method"
		vars := strings.Split(r.URL.Path, "/")
		fmt.Printf("ValidateMetricURL: vars=%v len=%d\n", vars, len(vars))
		// wrong type
		if len(vars) > 2 && vars[2] != models.Gauge && vars[2] != models.Counter {
			fmt.Println(vars[2] != models.Gauge)
			fmt.Println(vars[2] != models.Counter)
			fmt.Printf("ValidateMetricURL: invalid metric type %s\n", vars[2])
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// missing name
		if len(vars) < 4 || len(vars[3]) == 0 {
			fmt.Printf("ValidateMetricURL: missing metric name\n")
			w.WriteHeader(http.StatusNotFound)
			return
		}
		// invalid parameters count
		if len(vars) != 5 {
			// invalid parameters count
			fmt.Printf("ValidateMetricURL: bad request\n")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// invalid gauge value
		if _, err := strconv.ParseFloat(vars[4], 64); err != nil && vars[2] == models.Gauge {
			fmt.Printf("ValidateMetricURL: wrong gauge value %s\n", vars[4])
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// invalid counter increment
		if _, err := strconv.ParseInt(vars[4], 10, 64); err != nil && vars[2] == models.Counter {
			fmt.Printf("ValidateMetricURL: wrong counter increment %s\n", vars[4])
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		fmt.Printf("ValidateMetricURL: request is valid\n")
		next.ServeHTTP(w, r)
	})

}

func (srv *MetricsServer) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("UpdateHandler:  URL=%v\n", r.URL)
	if r.Method != http.MethodPost {
		fmt.Printf("UpdateHandler: wrong method %s\n", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	path, _ := strings.CutPrefix(r.URL.Path, "/update/")
	vars := strings.Split(path, "/")
	switch vars[0] {
	case models.Gauge:
		fmt.Printf("UpdateHandler: gauge with increment %s\n", vars[2])
		v, _ := strconv.ParseFloat(vars[2], 64)
		srv.storage.GaugeUpdate(vars[1], v)
	case models.Counter:
		fmt.Printf("UpdateHandler: counter with increment %s\n", vars[2])
		i, _ := strconv.ParseInt(vars[2], 10, 64)
		srv.storage.CounterUpdate(vars[1], i)
	}
	w.WriteHeader(http.StatusOK)
}
