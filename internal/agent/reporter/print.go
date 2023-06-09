package reporter

import (
	"fmt"

	"github.com/freepaddler/yap-metrics/internal/models"
	"github.com/freepaddler/yap-metrics/internal/store"
)

// PrintReporter is a test reporter to stdout
type PrintReporter struct {
	s store.Storage
}

func NewPrintReporter(s store.Storage) *PrintReporter {
	return &PrintReporter{
		s: s,
	}
}

func (r PrintReporter) Report() {
	m := r.s.GetAllMetrics()
	for _, v := range m {
		switch v.Type {
		case models.Counter:
			fmt.Printf("Metric: %s, type: %s, value: %d\n", v.Name, v.Type, v.IValue)
			r.s.DelCounter(v.Name)
		case models.Gauge:
			fmt.Printf("Metric: %s, type: %s, value: %f\n", v.Name, v.Type, v.FValue)
			r.s.DelGauge(v.Name)
		}
	}

}
