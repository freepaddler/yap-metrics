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

func PrepareTestStorage() (*Store, []gauge, []counter) {
	s := New()
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

func Test_IncCounter(t *testing.T) {
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
	s := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := s.IncCounter(tt.mName, tt.iValue)
			assert.Equal(t, tt.wantValue, v)
		})
	}
}

func Test_GetCounter(t *testing.T) {
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
				assert.Equal(t, tt.wantValue, v)
			}

		})
	}
}

func Test_DelCounter(t *testing.T) {
	s, _, _ := PrepareTestStorage()
	// check that counter exists in storage before deletion
	_, ok := s.GetCounter(eCounter2)
	require.Truef(t, ok, "Prepared set failed, counter should exist")
	// check that counter deleted form storage
	s.DelCounter(eCounter2)
	_, ok = s.GetCounter(eCounter2)
	assert.Falsef(t, ok, "counter exists, but should be deleted")
}

func Test_SetGauge(t *testing.T) {
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
	s := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.SetGauge(tt.mName, tt.fValue)
			v, ok := s.GetGauge(tt.mName)
			require.Equal(t, tt.wantOk, ok)
			assert.Equal(t, tt.wantValue, v)
		})
	}
}

func Test_GetGauge(t *testing.T) {
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
				assert.Equal(t, tt.wantValue, v)
			}

		})
	}
}

func Test_DelGauge(t *testing.T) {
	s, _, _ := PrepareTestStorage()
	// check that gauge exists in storage before deletion
	_, ok := s.GetGauge(eGauge2)
	require.Truef(t, ok, "Prepared set failed, gauge should exist")
	// check that gauge deleted form storage
	s.DelGauge(eGauge2)
	_, ok = s.GetGauge(eGauge2)
	assert.Falsef(t, ok, "gauge exists, but should be deleted")
}

func Test_Snapshot(t *testing.T) {
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
		require.ElementsMatch(t, m, m2, "snapshots without flush should match")
	})
	t.Run("With flush and empty", func(t *testing.T) {
		// storage snapshot without flush
		m := s.Snapshot(false)
		// storage snapshot with flush
		m1 := s.Snapshot(true)
		// should return same
		require.ElementsMatch(t, m, m1, "snapshots with and without flush should match")
		// storage snapshot after flush should be empty
		m2 := s.Snapshot(false)
		require.Equal(t, 0, len(m2), "no metrics should be returned")

	})
}
