package controller

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/freepaddler/yap-metrics/internal/pkg/models"
	"github.com/freepaddler/yap-metrics/mocks"
)

func Test_GetAll(t *testing.T) {
	var mockController = gomock.NewController(t)
	defer mockController.Finish()
	m := mocks.NewMockMemoryStore(mockController)

	c := New(m)

	m.EXPECT().Snapshot(false).Times(1)
	c.GetAll()
}

func Test_GetOne(t *testing.T) {
	var mockController = gomock.NewController(t)
	defer mockController.Finish()
	m := mocks.NewMockMemoryStore(mockController)

	c := New(m)

	tests := []struct {
		name            string
		req             models.MetricRequest
		wantErr         error
		wantInt         int64
		wantFloat       float64
		wantFound       bool
		wantCounterCall int
		wantGaugeCall   int
	}{
		{
			name: "counter found",
			req: models.MetricRequest{
				Name: "c1",
				Type: "counter",
			},
			wantInt:         int64(23),
			wantFound:       true,
			wantCounterCall: 1,
		},
		{
			name: "counter not found",
			req: models.MetricRequest{
				Name: "c1",
				Type: "counter",
			},
			wantFound:       false,
			wantCounterCall: 1,
			wantErr:         ErrMetricNotFound,
		},
		{
			name: "gauge found",
			req: models.MetricRequest{
				Name: "g1",
				Type: "gauge",
			},
			wantFloat:     -117.3,
			wantFound:     true,
			wantGaugeCall: 1,
		},

		{
			name: "gauge not found",
			req: models.MetricRequest{
				Name: "g1",
				Type: "gauge",
			},
			wantFound:     false,
			wantGaugeCall: 1,
			wantErr:       ErrMetricNotFound,
		},
		{
			name: "invalid type",
			req: models.MetricRequest{
				Name: "g1",
				Type: "gauge1",
			},
			wantErr: models.ErrInvalidMetric,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.EXPECT().GetCounter(tt.req.Name).Times(tt.wantCounterCall).Return(tt.wantInt, tt.wantFound)
			m.EXPECT().GetGauge(tt.req.Name).Times(tt.wantGaugeCall).Return(tt.wantFloat, tt.wantFound)
			got, err := c.GetOne(tt.req)
			if tt.wantErr != nil {
				require.True(t, errors.Is(err, tt.wantErr))
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.req.Name, got.Name)
				assert.Equal(t, tt.req.Type, got.Type)
				if tt.wantCounterCall > 0 {
					assert.Equal(t, tt.wantInt, *got.IValue)
				}
				if tt.wantGaugeCall > 0 {
					assert.Equal(t, tt.wantFloat, *got.FValue)
				}
			}
		})
	}
}

func Test_UpdateOne(t *testing.T) {
	var mockController = gomock.NewController(t)
	defer mockController.Finish()
	m := mocks.NewMockMemoryStore(mockController)

	c := New(m)

	tests := []struct {
		name        string
		mName       string
		mType       string
		wantErr     error
		wantCounter int
		wantGauge   int
		wantFloat   float64
		sendInt     int64
		wantInt     int64
	}{
		{
			name:        "counter",
			mName:       "c1",
			mType:       models.Counter,
			sendInt:     12,
			wantInt:     16,
			wantCounter: 1,
		},
		{
			name:      "gauge",
			mName:     "g1",
			mType:     models.Gauge,
			wantFloat: 117,
			wantGauge: 1,
		},
		{
			name:    "invalid type",
			mName:   "g1",
			mType:   "fakemetric",
			wantErr: models.ErrInvalidMetric,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metric := models.Metrics{Name: tt.mName, Type: tt.mType, IValue: &tt.sendInt, FValue: &tt.wantFloat}
			m.EXPECT().IncCounter(tt.mName, tt.sendInt).Times(tt.wantCounter).Return(tt.wantInt)
			m.EXPECT().SetGauge(tt.mName, tt.wantFloat).Times(tt.wantGauge).Return(tt.wantFloat)
			err := c.UpdateOne(&metric)
			if tt.wantErr != nil {
				require.True(t, errors.Is(err, tt.wantErr))
			} else {
				require.NoError(t, err)
				if tt.wantCounter > 0 {
					assert.Equal(t, tt.wantInt, *metric.IValue)
				}
				if tt.wantGauge > 0 {
					assert.Equal(t, tt.wantFloat, *metric.FValue)
				}
			}
		})
	}
}

func Test_UpdateMany(t *testing.T) {
	var mockController = gomock.NewController(t)
	defer mockController.Finish()
	m := mocks.NewMockMemoryStore(mockController)

	c := New(m)

	metrics := []models.Metrics{
		{
			Name:   "c1",
			Type:   models.Counter,
			IValue: new(int64),
		},
		{
			Name:   "c2",
			Type:   models.Counter,
			IValue: new(int64),
		},
		{
			Name:   "c1",
			Type:   models.Counter,
			IValue: new(int64),
		},
		{
			Name:   "g1",
			Type:   models.Gauge,
			FValue: new(float64),
		},
		{
			Name:   "g2",
			Type:   models.Gauge,
			FValue: new(float64),
		},
	}
	m.EXPECT().IncCounter(gomock.Any(), gomock.Any()).Times(3)
	m.EXPECT().SetGauge(gomock.Any(), gomock.Any()).Times(2)
	err := c.UpdateMany(metrics)
	require.NoError(t, err)
}
