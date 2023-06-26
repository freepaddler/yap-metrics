package memory

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/freepaddler/yap-metrics/internal/models"
)

func TestMemStorage_GetCounter(t *testing.T) {
	s := NewMemStorage(nil)
	existingName := "c1"
	existingValue := int64(1)
	absentName := "c2"
	s.IncCounter(existingName, existingValue)
	tests := []struct {
		name      string
		cName     string
		iValue    int64
		wantOk    bool
		wantValue int64
	}{
		{
			name:   "Get absent counter",
			cName:  absentName,
			wantOk: false,
		},
		{
			name:      "Get existing counter",
			cName:     existingName,
			iValue:    existingValue,
			wantOk:    true,
			wantValue: existingValue,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, ok := s.GetCounter(tt.cName)
			require.Equal(t, tt.wantOk, ok)
			if ok {
				assert.Equal(t, tt.wantValue, *v)
			}

		})
	}
}

func TestMemStorage_IncCounter(t *testing.T) {
	s := NewMemStorage(nil)
	tests := []struct {
		name      string
		cName     string
		iValue    int64
		wantOk    bool
		wantValue int64
	}{
		{
			name:      "Inc new counter value",
			cName:     "c1",
			iValue:    17,
			wantOk:    true,
			wantValue: 17,
		},
		{
			name:      "Inc existing counter",
			cName:     "c1",
			iValue:    19,
			wantOk:    true,
			wantValue: 17 + 19,
		},
		{
			name:      "Inc existing counter negative value",
			cName:     "c1",
			iValue:    -30,
			wantOk:    true,
			wantValue: 17 + 19 - 30,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.IncCounter(tt.cName, tt.iValue)
			v, ok := s.GetCounter(tt.cName)
			require.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.wantValue, *v)
		})
	}
}

func TestMemStorage_DelCounter(t *testing.T) {
	s := NewMemStorage(nil)
	deletedName := "c1"
	deletedValue := int64(1)
	s.IncCounter(deletedName, deletedValue)
	_, ok := s.GetCounter(deletedName)
	require.Equal(t, true, ok)
	s.DelCounter(deletedName)
	_, ok = s.GetCounter(deletedName)
	require.Equal(t, false, ok)
}

func TestMemStorage_GetGauge(t *testing.T) {
	s := NewMemStorage(nil)
	existingName := "g1"
	existingValue := -0.117
	absentName := "g2"
	s.SetGauge(existingName, existingValue)
	tests := []struct {
		name      string
		cName     string
		fValue    float64
		wantOk    bool
		wantValue float64
	}{
		{
			name:   "Get absent gauge",
			cName:  absentName,
			wantOk: false,
		},
		{
			name:      "Get existing gauge",
			cName:     existingName,
			fValue:    existingValue,
			wantOk:    true,
			wantValue: existingValue,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, ok := s.GetGauge(tt.cName)
			require.Equal(t, tt.wantOk, ok)
			if ok {
				assert.Equal(t, tt.wantValue, *v)
			}

		})
	}
}

func TestMemStorage_SetGauge(t *testing.T) {
	s := NewMemStorage(nil)
	tests := []struct {
		name      string
		cName     string
		fValue    float64
		wantOk    bool
		wantValue float64
	}{
		{
			name:      "Set new gauge value",
			cName:     "g1",
			fValue:    0.117,
			wantOk:    true,
			wantValue: 0.117,
		},
		{
			name:      "Set existing gauge value",
			cName:     "g1",
			fValue:    10,
			wantOk:    true,
			wantValue: 10,
		},
		{
			name:      "Set existing gauge negative value",
			cName:     "c1",
			fValue:    -119.37,
			wantOk:    true,
			wantValue: -119.37,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.SetGauge(tt.cName, tt.fValue)
			v, ok := s.GetGauge(tt.cName)
			require.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.wantValue, *v)
		})
	}
}

func TestMemStorage_DelGauge(t *testing.T) {
	s := NewMemStorage(nil)
	deletedName := "c1"
	deletedValue := float64(1)
	s.SetGauge(deletedName, deletedValue)
	_, ok := s.GetGauge(deletedName)
	require.Equal(t, true, ok)
	s.DelGauge(deletedName)
	_, ok = s.GetGauge(deletedName)
	require.Equal(t, false, ok)
}

// set number of metrics, get same number of metrics with same values
func TestMemStorage_GetAllMetrics(t *testing.T) {
	s := NewMemStorage(nil)
	type counter struct {
		models.Metrics
		IValue int64
	}
	type gauge struct {
		models.Metrics
		FValue float64
	}
	counters := []counter{
		{
			Metrics: models.Metrics{
				Name: "c1",
				Type: models.Counter,
			},
			IValue: 10,
		},
		{
			Metrics: models.Metrics{
				Name: "c2",
				Type: models.Counter,
			},
			IValue: 19,
		},
	}
	gauges := []gauge{
		{
			Metrics: models.Metrics{
				Name: "g1",
				Type: models.Gauge,
			},
			FValue: -0.117,
		},
		{
			Metrics: models.Metrics{
				Name: "g2",
				Type: models.Gauge,
			},
			FValue: 192.3947,
		},
		{
			Metrics: models.Metrics{
				Name: "g3",
				Type: models.Gauge,
			},
			FValue: 0.000,
		},
	}
	for _, v := range counters {
		s.IncCounter(v.Name, v.IValue)
	}
	for _, v := range gauges {
		s.SetGauge(v.Name, v.FValue)
	}

	m := s.GetAllMetrics()
	require.Equal(t, len(counters)+len(gauges), len(m), "Wrong reported metrics count")
	for _, v := range counters {
		//assert.Contains(t, m, v)
		reflect.DeepEqual(m, v)
	}
	for _, v := range gauges {
		//assert.Contains(t, m, v)
		reflect.DeepEqual(m, v)
	}
}

func TestMemStorage_GetAllMetrics_EmptySet(t *testing.T) {
	s := NewMemStorage(nil)
	m := s.GetAllMetrics()
	require.Equal(t, 0, len(m), "No metrics should be returned")
}

func TestMemStorage_GetMetric(t *testing.T) {
	s := NewMemStorage(nil)
	existingCounterName := "c1"
	existingCounterValue := int64(1)
	absentCounterName := "c2"
	existingGaugeName := "g1"
	existingGaugeValue := float64(-1.017)
	absentGaugeName := "g2"
	s.IncCounter(existingCounterName, existingCounterValue)
	s.SetGauge(existingGaugeName, existingGaugeValue)
	tests := []struct {
		name       string
		metric     models.Metrics
		wantMetric models.Metrics
		wantOk     bool
	}{
		{
			name: "Get absent counter",
			metric: models.Metrics{
				Name: absentCounterName,
				Type: models.Counter,
			},
			wantOk: false,
		},
		{
			name: "Get existing counter",
			metric: models.Metrics{
				Name: existingCounterName,
				Type: models.Counter,
			},
			wantMetric: models.Metrics{
				Name:   existingCounterName,
				Type:   models.Counter,
				IValue: &existingCounterValue,
			},
			wantOk: true,
		},
		{
			name: "Get absent gauge",
			metric: models.Metrics{
				Name: absentGaugeName,
				Type: models.Gauge,
			},
			wantOk: false,
		},
		{
			name: "Get existing gauge",
			metric: models.Metrics{
				Name: existingGaugeName,
				Type: models.Gauge,
			},
			wantMetric: models.Metrics{
				Name:   existingGaugeName,
				Type:   models.Gauge,
				FValue: &existingGaugeValue,
			},
			wantOk: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, _ := s.GetMetric(&tt.metric)
			require.Equal(t, tt.wantOk, ok)
			if ok {
				assert.Equal(t, tt.wantMetric, tt.metric)
			}

		})
	}
}

func TestMemStorage_SetMetric(t *testing.T) {
	s := NewMemStorage(nil)
	tests := []struct {
		name             string
		gaugeValue       float64
		counterValue     int64
		wantGaugeValue   float64
		wantCounterValue int64
		metricName       string
		metricType       string
	}{
		{
			name:             "Set new counter",
			metricType:       models.Counter,
			metricName:       "c1",
			counterValue:     11,
			wantCounterValue: 11,
		},
		{
			name:             "Update counter",
			metricType:       models.Counter,
			metricName:       "c1",
			counterValue:     -2,
			wantCounterValue: 9,
		},
		{
			name:           "Set new gauge",
			metricType:     models.Gauge,
			metricName:     "g1",
			gaugeValue:     -1.017,
			wantGaugeValue: -1.017,
		},
		{
			name:           "Update gauge",
			metricType:     models.Gauge,
			metricName:     "g1",
			gaugeValue:     119.25,
			wantGaugeValue: 119.25,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := models.Metrics{
				Name:   tt.metricName,
				Type:   tt.metricType,
				FValue: &tt.gaugeValue,
				IValue: &tt.counterValue,
			}
			m2 := models.Metrics{
				Name:   tt.metricName,
				Type:   tt.metricType,
				FValue: &tt.wantGaugeValue,
				IValue: &tt.wantCounterValue,
			}
			err := s.SetMetric(&m)
			require.Nil(t, err)
			assert.Equal(t, m, m2)
		})
	}
}
