package models

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetrics_StringVal(t *testing.T) {
	tests := []struct {
		name   string
		mName  string
		mType  string
		iValue int64
		fValue float64
		want   string
	}{
		{
			name:   "counter",
			mName:  "c1",
			mType:  Counter,
			iValue: -14,
			fValue: 0,
			want:   "-14",
		},
		{
			name:   "gauge",
			mName:  "g1",
			mType:  Gauge,
			fValue: -0.119010,
			want:   "-0.11901",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metrics{
				Name:   tt.mName,
				Type:   tt.mType,
				FValue: &tt.fValue,
				IValue: &tt.iValue,
			}
			if got := m.StringVal(); got != tt.want {
				t.Errorf("StringVal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMetrics_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		raw      []byte
		wantName string
		wantType string
		wantVal  string
		wantErr  error
	}{
		{
			name:     "counter with value",
			raw:      []byte(`{"id":"m","type":"counter","delta":10}`),
			wantName: "m",
			wantType: Counter,
			wantVal:  "10",
			wantErr:  nil,
		},
		{
			name:     "counter without value",
			raw:      []byte(`{"id":"m","type":"counter"}`),
			wantName: "m",
			wantType: Counter,
			wantVal:  "",
			wantErr:  nil,
		},
		{
			name:     "gauge with value",
			raw:      []byte(`{"id":"m","type":"gauge","value":-0.119}`),
			wantName: "m",
			wantType: Gauge,
			wantVal:  "-0.119",
			wantErr:  nil,
		},
		{
			name:     "gauge without value",
			raw:      []byte(`{"id":"m","type":"gauge"}`),
			wantName: "m",
			wantType: Gauge,
			wantVal:  "",
			wantErr:  nil,
		},

		{
			name:    "no name",
			raw:     []byte(`{"type":"counter","delta":10}`),
			wantErr: ErrBadMetric,
		},
		{
			name:    "invalid type",
			raw:     []byte(`{"id":"m","type":"counter1","delta":10}`),
			wantErr: ErrBadMetric,
		},
		{
			name:    "counter with value",
			raw:     []byte(`{"id":"m","type":"counter","value":-0.117}`),
			wantErr: ErrBadMetric,
		},
		{
			name:    "gauge with delta",
			raw:     []byte(`{"id":"m","type":"gauge","delta":10}`),
			wantErr: ErrBadMetric,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var m Metrics
			err := json.Unmarshal(tt.raw, &m)
			if tt.wantErr != nil {
				assert.Truef(t, errors.Is(err, tt.wantErr), "Expect error '%v', got '%v'", tt.wantErr, err)
			} else {
				require.NoError(t, err, "Expect no error, got:", err)
				assert.Equalf(t, tt.wantName, m.Name, "Expect name %s, got %s", tt.wantName, m.Name)
				assert.Equalf(t, tt.wantType, m.Type, "Expect type %s, got %s", tt.wantType, m.Type)
				assert.Equalf(t, tt.wantVal, m.StringVal(), "Expect name %s, got %s", tt.wantVal, m.StringVal())
			}

		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		n       string
		t       string
		v       string
		wantV   string
		wantErr error
	}{
		{
			name:    "counter with value",
			n:       "m",
			t:       Counter,
			v:       "10",
			wantV:   "10",
			wantErr: nil,
		},
		{
			name:    "counter without value",
			n:       "m",
			t:       Counter,
			v:       "",
			wantV:   "",
			wantErr: nil,
		},
		{
			name:    "gauge with value",
			n:       "m",
			t:       Gauge,
			v:       "1.237",
			wantV:   "1.237",
			wantErr: nil,
		},
		{
			name:    "gauge without value",
			n:       "m",
			t:       Gauge,
			v:       "",
			wantV:   "",
			wantErr: nil,
		},
		{
			name:    "no name",
			wantErr: ErrBadMetric,
		},
		{
			name:    "invalid type",
			n:       "m",
			t:       "sometype",
			wantErr: ErrBadMetric,
		},
		{
			name:    "counter with non-int",
			n:       "m",
			t:       Counter,
			v:       "0.13",
			wantErr: ErrBadMetric,
		},
		{
			name:    "gauge with non-float",
			n:       "m",
			t:       Gauge,
			v:       "qwe0.13",
			wantErr: ErrBadMetric,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotM, err := NewMetric(tt.n, tt.t, tt.v)
			if tt.wantErr != nil {
				assert.Truef(t, errors.Is(err, tt.wantErr), "Expect error '%v', got '%v'", tt.wantErr, err)
			} else {
				assert.Equal(t, tt.wantV, gotM.StringVal())
			}
		})
	}
}
