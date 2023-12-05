package proto

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/freepaddler/yap-metrics/internal/pkg/models"
)

func pointer[T any](val T) *T {
	return &val
}

func TestMetric_FromMetrics(t *testing.T) {
	tests := []struct {
		name    string
		metric  models.Metrics
		want    *Metric
		wantErr error
	}{
		{
			name:   "counter",
			metric: models.Metrics{Name: "c1", Type: models.Counter, IValue: pointer(int64(-19))},
			want:   &Metric{Id: "c1", Type: Metric_COUNTER, Delta: int64(-19)},
		},
		{
			name:   "gauge",
			metric: models.Metrics{Name: "g1", Type: models.Gauge, FValue: pointer(12312.1312)},
			want:   &Metric{Id: "g1", Type: Metric_GAUGE, Value: 12312.1312},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x := &Metric{}
			err := x.FromMetrics(tt.metric)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, x)
			}
		})
	}
}

func TestMetric_ToMetrics(t *testing.T) {
	type fields struct {
		ID    string
		Type  Metric_Type
		Value float64
		Delta int64
	}
	tests := []struct {
		name    string
		fields  fields
		want    models.Metrics
		wantErr error
	}{
		{
			name:   "counter",
			fields: fields{ID: "c1", Type: Metric_COUNTER, Delta: 138},
			want:   models.Metrics{Name: "c1", Type: models.Counter, IValue: pointer(int64(138))},
		},
		{
			name:   "gauge",
			fields: fields{ID: "g1", Type: Metric_GAUGE, Value: -0.117},
			want:   models.Metrics{Name: "g1", Type: models.Gauge, FValue: pointer(-0.117)},
		},
		{
			name:    "no name",
			fields:  fields{Type: Metric_GAUGE, Value: -0.117},
			want:    models.Metrics{Name: "g1", Type: models.Gauge, FValue: pointer(-0.117)},
			wantErr: models.ErrInvalidMetric,
		},
		{
			name:    "bad type",
			fields:  fields{ID: "g1", Type: Metric_UNSPECIFIED, Value: -0.117},
			want:    models.Metrics{Name: "g1", Type: models.Gauge, FValue: pointer(-0.117)},
			wantErr: models.ErrInvalidMetric,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x := &Metric{
				Id:    tt.fields.ID,
				Type:  tt.fields.Type,
				Value: tt.fields.Value,
				Delta: tt.fields.Delta,
			}
			got, err := x.ToMetrics()
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, got)
			}
		})
	}
}
