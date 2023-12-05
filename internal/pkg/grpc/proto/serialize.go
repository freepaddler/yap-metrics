package proto

//go:generate mockgen -source metrics_grpc.pb.go -package=mocks -destination ../../../../mocks/MetricsGRPCServer_mock.go

import (
	"github.com/freepaddler/yap-metrics/internal/pkg/models"
)

// ToMetrics returns models.Metrics struct from protobuf Metric
func (x *Metric) ToMetrics() (models.Metrics, error) {
	var m models.Metrics
	if x.Id == "" {
		return m, models.ErrInvalidName
	}
	m.Name = x.Id
	switch x.Type {
	case Metric_COUNTER:
		m.Type = models.Counter
		m.IValue = &x.Delta
	case Metric_GAUGE:
		m.Type = models.Gauge
		m.FValue = &x.Value
	default:
		return m, models.ErrInvalidType
	}
	return m, nil
}

// FromMetrics assigns models.Metrics to protobuf Metric
func (x *Metric) FromMetrics(m models.Metrics) error {
	x.Id = m.Name
	switch m.Type {
	case models.Counter:
		x.Type = Metric_COUNTER
		x.Delta = *m.IValue
	case models.Gauge:
		x.Type = Metric_GAUGE
		x.Value = *m.FValue
	default:
		return models.ErrInvalidMetric
	}
	return nil
}
