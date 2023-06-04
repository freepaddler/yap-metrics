package agent

import (
	"fmt"

	"github.com/freepaddler/yap-metrics/internal/models"
)

//type HttpReporter struct {
//	host string
//	port string
//	Reporter
//}
//
//func NewHttpReporter(host, port string) *HttpReporter {
//	return &HttpReporter{
//		host: host,
//		port: port,
//	}
//}
//
//func (r HttpReporter) Report(m models.Metrics) bool {
//	return true
//}

// PrintReporter is a test reporter to stdout
type PrintReporter struct {
	Reporter
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
