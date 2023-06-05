package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/freepaddler/yap-metrics/internal/models"
)

// UpdateHandler validates update request and writes metrics to storage
func (srv *MetricsServer) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("UpdateHandler: Request received  URL=%v\n", r.URL)
	// TODO: should be checked in third-party router
	if r.Method != http.MethodPost {
		// curl -i http://localhost:8080/update/bla...
		fmt.Printf("UpdateHandler: wrong method %s\n", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// TODO: change to third-party router to simplify validation
	//remove request prefix
	path, _ := strings.CutPrefix(r.URL.Path, "/update/")
	// split url string to parts
	// 0 = type, 1 = name, 2 = value
	vars := strings.Split(path, "/")
	// TODO: simple way is to create structure and use json.Unmarshal for validation or use map
	switch {

	// check metric type
	// curl -X POST -i http://localhost:8080/update/something
	// curl -X POST -i http://localhost:8080/update/
	case vars[0] != models.Gauge && vars[0] != models.Counter:
		fmt.Printf("UpdateHandler: wrong metric type '%s'\n", vars[0])
		w.WriteHeader(http.StatusBadRequest)
		return

	// check metric name
	// curl -X POST -i http://localhost:8080/update/counter
	// curl -X POST -i http://localhost:8080/update/gauge/
	case len(vars) < 2 || len(vars[1]) == 0:
		fmt.Printf("UpdateHandler: missing metric name \n")
		w.WriteHeader(http.StatusNotFound)
		return

	// curl -X POST -i http://localhost:8080/update/counter/c1
	case len(vars) < 3:
		fmt.Printf("UpdateHandler: missing metric value\n")
		w.WriteHeader(http.StatusBadRequest)
		return

	// validate values
	// curl -X POST -i http://localhost:8080/update/counter/c1/10.002
	// curl -X POST -i http://localhost:8080/update/gauge/g1/none
	// curl -X POST -i http://localhost:8080/update/gauge/g1/
	// curl -X POST -i http://localhost:8080/update/gauge/g1/-1.75
	// curl -X POST -i http://localhost:8080/update/gauge/g1/1.0
	// curl -X POST -i http://localhost:8080/update/counter/c1/10
	// curl -X POST -i http://localhost:8080/update/counter/c1/20
	case len(vars) == 3:
		switch vars[0] {
		case models.Counter:
			if v, err := strconv.ParseInt(vars[2], 10, 64); err != nil {
				fmt.Printf("UpdateHandler: wrong counter increment '%s'\n", vars[2])
				w.WriteHeader(http.StatusBadRequest)
				return
			} else {
				srv.storage.CounterUpdate(vars[1], v)
			}
		case models.Gauge:
			if v, err := strconv.ParseFloat(vars[2], 64); err != nil {
				fmt.Printf("UpdateHandler: wrong gauge value '%s'\n", vars[2])
				w.WriteHeader(http.StatusBadRequest)
				return
			} else {
				srv.storage.GaugeUpdate(vars[1], v)
			}
		}

	// too long path
	// curl -X POST -i http://localhost:8080/update/counter/c1/10/20/30/40
	default:
		fmt.Printf("UpdateHandler: invalid request \n")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}
