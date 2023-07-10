package memory

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/freepaddler/yap-metrics/internal/pkg/models"
)

const (
	// existing metrics
	eCounter1    string  = "c1"
	eCounter1Val int64   = 10
	eCounter2    string  = "c2"
	eCounter2Val int64   = 19
	eGauge1      string  = "g1"
	eGauge1Val   float64 = -0.117
	eGauge2      string  = "g2"
	eGauge2Val   float64 = 192.3947
	eGauge3      string  = "g3"
	eGauge3Val   float64 = 0.000

	// new metrics
	newCounter    string  = "c0"
	newCounterVal int64   = 7
	newGauge      string  = "g0"
	newGaugeVal   float64 = 1.019
)

type counter struct {
	models.Metrics
	IValue int64
}

type gauge struct {
	models.Metrics
	FValue float64
}

func PrepareTestStorage() (*MemStorage, []gauge, []counter) {
	s := NewMemStorage()
	counters := []counter{
		{
			Metrics: models.Metrics{
				Name: eCounter1,
				Type: models.Counter,
			},
			IValue: eCounter1Val,
		},
		{
			Metrics: models.Metrics{
				Name: eCounter2,
				Type: models.Counter,
			},
			IValue: eCounter2Val,
		},
	}
	gauges := []gauge{
		{
			Metrics: models.Metrics{
				Name: eGauge1,
				Type: models.Gauge,
			},
			FValue: eGauge1Val,
		},
		{
			Metrics: models.Metrics{
				Name: eGauge2,
				Type: models.Gauge,
			},
			FValue: eGauge2Val,
		},
		{
			Metrics: models.Metrics{
				Name: eGauge3,
				Type: models.Gauge,
			},
			FValue: eGauge3Val,
		},
	}
	for _, v := range counters {
		s.IncCounter(v.Name, v.IValue)
	}
	for _, v := range gauges {
		s.SetGauge(v.Name, v.FValue)
	}
	return s, gauges, counters
}

func TestMemStorage_IncCounter_UpdateSingleMetric(t *testing.T) {
	tests := []struct {
		name      string
		mName     string
		iValue    int64
		wantOk    bool
		wantValue int64
	}{
		{
			name:      "Inc new counter value",
			mName:     newCounter,
			iValue:    newCounterVal,
			wantOk:    true,
			wantValue: newCounterVal,
		},
		{
			name:      "Inc existing counter",
			mName:     newCounter,
			iValue:    19,
			wantOk:    true,
			wantValue: newCounterVal + 19,
		},
		{
			name:      "Inc existing counter negative value",
			mName:     newCounter,
			iValue:    -30,
			wantOk:    true,
			wantValue: newCounterVal + 19 - 30,
		},
	}
	s := NewMemStorage()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := s.IncCounter(tt.mName, tt.iValue)
			assert.Equal(t, tt.wantValue, v)
		})
	}
	s = NewMemStorage()
	for _, tt := range tests {
		t.Run("Update "+tt.name+" (overwrite false)", func(t *testing.T) {
			m := models.Metrics{
				Name:   tt.mName,
				Type:   models.Counter,
				IValue: nil,
			}
			m.IValue = &tt.iValue

			invalid := s.UpdateMetrics([]models.Metrics{m}, false)
			require.Emptyf(t, invalid, "invalid records found, but should not")
			val, ok := s.GetCounter(tt.mName)
			require.Equal(t, tt.wantOk, ok)
			if ok {
				assert.Equal(t, tt.wantValue, *val)
			}
		})
	}
	s = NewMemStorage()
	for _, tt := range tests {
		t.Run("Update "+tt.name+" (overwrite true)", func(t *testing.T) {
			m := models.Metrics{
				Name:   tt.mName,
				Type:   models.Counter,
				IValue: nil,
			}
			m.IValue = &tt.iValue

			invalid := s.UpdateMetrics([]models.Metrics{m}, false)
			require.Emptyf(t, invalid, "invalid records found, but should not")
			val, ok := s.GetCounter(tt.mName)
			require.Equal(t, tt.wantOk, ok)
			if ok {
				assert.Equal(t, tt.iValue, *val)
			}
		})
	}
}
func TestMemStorage_GetCounter_GetMetric(t *testing.T) {
	s, _, _ := PrepareTestStorage()
	tests := []struct {
		name      string
		mName     string
		wantOk    bool
		wantValue int64
	}{
		{
			name:   "Get absent counter",
			mName:  newCounter,
			wantOk: false,
		},
		{
			name:      "Get existing counter",
			mName:     eCounter1,
			wantOk:    true,
			wantValue: eCounter1Val,
		},
	}
	for _, tt := range tests {
		t.Run("GetCounter: "+tt.name, func(t *testing.T) {
			v, ok := s.GetCounter(tt.mName)
			require.Equal(t, tt.wantOk, ok)
			if ok {
				assert.Equal(t, tt.wantValue, *v)
			}

		})
	}
	for _, tt := range tests {
		t.Run("GetMetric: "+tt.name, func(t *testing.T) {
			m := models.Metrics{
				Name: tt.mName,
				Type: models.Counter,
			}
			ok, _ := s.GetMetric(&m)
			//require.NoError(t, err)
			require.Equal(t, tt.wantOk, ok)
			if ok {
				assert.Equal(t, tt.wantValue, *m.IValue)
			}

		})
	}
}
func TestMemStorage_DelCounter(t *testing.T) {
	s, _, _ := PrepareTestStorage()
	// check that counter exists in storage before deletion
	_, ok := s.GetCounter(eCounter2)
	require.Truef(t, ok, "Prepared set failed, counter should exist")
	// check that counter deleted form storage
	s.DelCounter(eCounter2)
	_, ok = s.GetCounter(eCounter2)
	assert.Falsef(t, ok, "counter exists, but should be deleted")
}

func TestMemStorage_SetGauge_UpdateSingleMetric(t *testing.T) {
	tests := []struct {
		name      string
		mName     string
		fValue    float64
		wantOk    bool
		wantValue float64
	}{
		{
			name:      "Set new gauge value",
			mName:     newGauge,
			fValue:    newGaugeVal,
			wantOk:    true,
			wantValue: newGaugeVal,
		},
		{
			name:      "Set existing gauge value",
			mName:     newGauge,
			fValue:    10,
			wantOk:    true,
			wantValue: 10,
		},
		{
			name:      "Set existing gauge negative value",
			mName:     newGauge,
			fValue:    -119.37,
			wantOk:    true,
			wantValue: -119.37,
		},
	}
	s := NewMemStorage()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.SetGauge(tt.mName, tt.fValue)
			v, ok := s.GetGauge(tt.mName)
			require.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.wantValue, *v)
		})
	}
	s = NewMemStorage()
	for _, tt := range tests {
		t.Run(tt.name+"(overwrite false)", func(t *testing.T) {
			m := models.Metrics{
				Name:   tt.mName,
				Type:   models.Gauge,
				FValue: nil,
			}
			m.FValue = &tt.fValue

			invalid := s.UpdateMetrics([]models.Metrics{m}, false)
			require.Emptyf(t, invalid, "invalid records found, but should not")
			val, ok := s.GetGauge(tt.mName)
			require.Equal(t, tt.wantOk, ok)
			if ok {
				assert.Equal(t, tt.wantValue, *val)
			}
		})
	}
	s = NewMemStorage()
	for _, tt := range tests {
		t.Run(tt.name+"(overwrite true)", func(t *testing.T) {
			m := models.Metrics{
				Name:   tt.mName,
				Type:   models.Gauge,
				FValue: nil,
			}
			m.FValue = &tt.fValue

			invalid := s.UpdateMetrics([]models.Metrics{m}, true)
			require.Emptyf(t, invalid, "invalid records found, but should not")
			val, ok := s.GetGauge(tt.mName)
			require.Equal(t, tt.wantOk, ok)
			if ok {
				assert.Equal(t, tt.wantValue, *val)
			}
		})
	}
}
func TestMemStorage_GetGauge_GetMetric(t *testing.T) {
	s, _, _ := PrepareTestStorage()
	tests := []struct {
		name      string
		mName     string
		wantOk    bool
		wantValue float64
	}{
		{
			name:   "Get absent gauge",
			mName:  newGauge,
			wantOk: false,
		},
		{
			name:      "Get existing gauge",
			mName:     eGauge1,
			wantOk:    true,
			wantValue: eGauge1Val,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, ok := s.GetGauge(tt.mName)
			require.Equal(t, tt.wantOk, ok)
			if ok {
				assert.Equal(t, tt.wantValue, *v)
			}

		})
	}
	for _, tt := range tests {
		t.Run("GetMetric: "+tt.name, func(t *testing.T) {
			m := models.Metrics{
				Name: tt.mName,
				Type: models.Gauge,
			}
			ok, _ := s.GetMetric(&m)
			//require.NoError(t, err)
			require.Equal(t, tt.wantOk, ok)
			if ok {
				assert.Equal(t, tt.wantValue, *m.FValue)
			}

		})
	}
}
func TestMemStorage_DelGauge(t *testing.T) {
	s, _, _ := PrepareTestStorage()
	// check that gauge exists in storage before deletion
	_, ok := s.GetGauge(eGauge2)
	require.Truef(t, ok, "Prepared set failed, gauge should exist")
	// check that gauge deleted form storage
	s.DelGauge(eGauge2)
	_, ok = s.GetGauge(eGauge2)
	assert.Falsef(t, ok, "gauge exists, but should be deleted")
}

func TestMemStorage_Snapshot(t *testing.T) {
	s, gauges, counters := PrepareTestStorage()

	t.Run("Without flush", func(t *testing.T) {
		m := s.Snapshot(false)
		require.Equal(t, len(gauges)+len(counters), len(m), "Wrong reported metrics count")
		for _, v := range counters {
			reflect.DeepEqual(m, v)
		}
		for _, v := range gauges {
			reflect.DeepEqual(m, v)
		}
		m2 := s.Snapshot(false)
		// flush false, so slices must be equal
		require.Equal(t, m, m2)
	})
	t.Run("With flush and empty", func(t *testing.T) {
		// storage snapshot without flush
		m := s.Snapshot(false)
		// storage snapshot with flush
		m1 := s.Snapshot(true)
		// should return same
		require.Equal(t, m, m1, "snapshots with and without flush should match")
		// storage snapshot after flush should be empty
		m2 := s.Snapshot(false)
		require.Equal(t, 0, len(m2), "no metrics should be returned")

	})
}
func TestMemStorage_UpdateMultiple_and_Hook(t *testing.T) {

	// new metrics for insert
	mCounterN := models.Metrics{
		Name:   newCounter,
		Type:   models.Counter,
		IValue: new(int64),
	}
	*mCounterN.IValue = newCounterVal
	wantCounterN := newCounterVal
	mGaugeN := models.Metrics{
		Name:   newGauge,
		Type:   models.Gauge,
		FValue: new(float64),
	}
	*mGaugeN.FValue = newGaugeVal
	wantGaugeN := newGaugeVal
	// existing metrics to update
	mCounterE := models.Metrics{
		Name:   eCounter2,
		Type:   models.Counter,
		IValue: new(int64),
	}
	*mCounterE.IValue = eCounter2Val
	wantCounterE := eCounter2Val * 2
	mGaugeE := models.Metrics{
		Name:   eGauge1,
		Type:   models.Gauge,
		FValue: new(float64),
	}
	*mGaugeE.FValue = eGauge1Val
	wantGaugeE := eGauge1Val
	// invalid metrics
	mInv1 := models.Metrics{
		Name: "invalid1",
		Type: "someType",
	}
	mInv2 := models.Metrics{
		Name: "invalid2",
		Type: "someType2",
	}
	// check invalid
	wantInvalid := []models.Metrics{mInv1, mInv2}
	s, _, _ := PrepareTestStorage()
	globalHookMetricsVar = make([]models.Metrics, 0)
	s.RegisterHooks(testHook)
	invalid := s.UpdateMetrics([]models.Metrics{mCounterN, mGaugeN, mInv1, mCounterE, mGaugeE, mInv2}, false)
	t.Run("invalid metrics list", func(t *testing.T) {
		assert.ElementsMatch(t, wantInvalid, invalid)
	})
	t.Run("update success", func(t *testing.T) {
		assert.Equal(t, wantInvalid, invalid)
		getCounterN, ok := s.GetCounter(mCounterN.Name)
		require.True(t, ok)
		assert.Equal(t, wantCounterN, *getCounterN)
		getCounterE, ok := s.GetCounter(mCounterE.Name)
		require.True(t, ok)
		assert.Equal(t, wantCounterE, *getCounterE)
		getGaugeN, ok := s.GetGauge(mGaugeN.Name)
		require.True(t, ok)
		assert.Equal(t, wantGaugeN, *getGaugeN)
		getGaugeE, ok := s.GetGauge(mGaugeE.Name)
		require.True(t, ok)
		assert.Equal(t, wantGaugeE, *getGaugeE)
	})
	t.Run("hook called", func(t *testing.T) {
		wantValid := []models.Metrics{mCounterN, mGaugeN, mCounterE, mGaugeE}
		assert.ElementsMatch(t, wantValid, globalHookMetricsVar)
	})

}

var globalHookMetricsVar []models.Metrics

func testHook(m []models.Metrics) {
	globalHookMetricsVar = m
}
