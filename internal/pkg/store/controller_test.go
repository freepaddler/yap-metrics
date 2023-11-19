package store

import (
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/freepaddler/yap-metrics/internal/pkg/models"
	"github.com/freepaddler/yap-metrics/mocks"
)

func TestMetricsController_CollectCounter(t *testing.T) {
	var mockController = gomock.NewController(t)
	defer mockController.Finish()
	m := mocks.NewMockStore(mockController)

	name := "counter1"
	value := int64(-12)

	c := NewStorageController(m)

	m.EXPECT().IncCounter(name, value).Times(1)
	c.CollectCounter(name, value)
}

func TestMetricsController_CollectGauge(t *testing.T) {
	var mockController = gomock.NewController(t)
	defer mockController.Finish()
	m := mocks.NewMockStore(mockController)

	name := "gauge1"
	value := -0.117

	c := NewStorageController(m)

	m.EXPECT().SetGauge(name, value).Times(2)
	tStart := time.Now()
	c.CollectGauge(name, value)
	require.WithinRange(t, c.gaugesTS[name], tStart, time.Now(), "Invalid timestamp in map")
	tStart = time.Now()
	c.CollectGauge(name, value)
	require.WithinRange(t, c.gaugesTS[name], tStart, time.Now(), "Invalid timestamp in map")
}

func TestMetricsController_ReportAll(t *testing.T) {
	var mockController = gomock.NewController(t)
	defer mockController.Finish()
	m := mocks.NewMockStore(mockController)

	c := NewStorageController(m)

	m.EXPECT().Snapshot(true).Times(1)
	tStart := time.Now()
	_, ts := c.ReportAll()
	require.WithinRange(t, ts, tStart, time.Now(), "Unexpected report timestamp")
}

func TestMetricsController_RestoreReport(t *testing.T) {
	var mockController = gomock.NewController(t)
	defer mockController.Finish()
	m := mocks.NewMockStore(mockController)

	// counter in report
	cVal := int64(12)
	counter := models.Metrics{
		Name:   "c1",
		Type:   models.Counter,
		IValue: &cVal,
	}
	// gauge in report
	gVal := 1.119
	gauge := models.Metrics{
		Name:   "g1",
		Type:   models.Gauge,
		FValue: &gVal,
	}
	// report
	report := []models.Metrics{counter, gauge}

	t.Run("Restore to empty store", func(t *testing.T) {
		c := NewStorageController(m)
		ts := time.Now().Add(-1 * time.Second)
		// both metrics should be updated
		m.EXPECT().IncCounter(counter.Name, *counter.IValue).Times(1)
		m.EXPECT().SetGauge(gauge.Name, *gauge.FValue).Times(1)
		c.RestoreLatest(report, ts)
		require.Equal(t, ts, c.gaugesTS[gauge.Name], "Expect '%t' in gaugesTs map, got '%t'", ts, c.gaugesTS[gauge.Name])
	})

	t.Run("Restore to updated store", func(t *testing.T) {
		c := NewStorageController(m)

		m.EXPECT().Snapshot(true).Return(report).Times(1)
		r, reportTS := c.ReportAll()

		m.EXPECT().IncCounter(gomock.Any(), gomock.Any())
		m.EXPECT().SetGauge(gomock.Any(), gomock.Any())
		c.CollectGauge("g1", 1)
		c.CollectCounter("c1", 1)

		// only counter should be updated
		m.EXPECT().IncCounter(counter.Name, *counter.IValue).Times(1)
		c.RestoreLatest(r, reportTS)

		// gauge timestamp should not be changed
		require.True(t, c.gaugesTS[gauge.Name].After(reportTS), "Expect time in gaugesTs map later than report time")
	})
}

func Test_GetAll(t *testing.T) {
	var mockController = gomock.NewController(t)
	defer mockController.Finish()
	m := mocks.NewMockStore(mockController)

	c := NewStorageController(m)

	m.EXPECT().Snapshot(false).Times(1)
	c.GetAll()
}

func Test_GetOne(t *testing.T) {
	var mockController = gomock.NewController(t)
	defer mockController.Finish()
	m := mocks.NewMockStore(mockController)

	c := NewStorageController(m)

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
	m := mocks.NewMockStore(mockController)

	c := NewStorageController(m)

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
	m := mocks.NewMockStore(mockController)

	c := NewStorageController(m)

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

func TestController_Ping(t *testing.T) {
	var mockController = gomock.NewController(t)
	defer mockController.Finish()
	m := mocks.NewMockStore(mockController)

	c := NewStorageController(m)
	m.EXPECT().Ping().Times(1).Return(nil)
	err := c.Ping()
	require.NoError(t, err)
	m.EXPECT().Ping().Times(1).Return(errors.New("some err"))
	err = c.Ping()
	require.Error(t, err)

}
