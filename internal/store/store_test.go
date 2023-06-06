package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemStorage_CounterUpdate(t *testing.T) {
	tests := []struct {
		name        string
		metricName  string
		increment   int64
		want        int64
		metricStore map[string]int64
	}{
		{
			name:        "add increment to new counter",
			metricName:  "count1",
			increment:   10,
			want:        10,
			metricStore: map[string]int64{},
		},
		{
			name:        "add increment to existing counter",
			metricName:  "count1",
			increment:   -10,
			want:        2,
			metricStore: map[string]int64{"count1": 12},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				counters: tt.metricStore,
			}
			require.NotPanics(t, func() { ms.CounterSet(tt.metricName, tt.increment) })
			require.Contains(t, ms.counters, tt.metricName)
			assert.Equal(t, tt.want, ms.counters[tt.metricName])
		})
	}
}

func TestMemStorage_GaugeUpdate(t *testing.T) {
	tests := []struct {
		name        string
		metricName  string
		value       float64
		want        float64
		metricStore map[string]float64
	}{
		{
			name:        "place value to new gauge",
			metricName:  "gauge1",
			value:       0.01,
			want:        0.01,
			metricStore: map[string]float64{},
		},
		{
			name:        "place value to existing gauge",
			metricName:  "gauge1",
			value:       -117.25,
			want:        -117.25,
			metricStore: map[string]float64{"gauge1": 1000},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				gauges: tt.metricStore,
			}
			require.NotPanics(t, func() { ms.GaugeSet(tt.metricName, tt.value) })
			require.Contains(t, ms.gauges, tt.metricName)
			assert.Equal(t, tt.want, ms.gauges[tt.metricName])
		})
	}
}
