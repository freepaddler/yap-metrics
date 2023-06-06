package agent

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/freepaddler/yap-metrics/internal/models"
)

// check counter increment (should we?)
func Test_counter_Inc(t *testing.T) {
	type fields struct {
		prev  int64
		value int64
	}
	tests := []struct {
		name      string
		fields    fields
		increment int64
		want      *counter
	}{
		{
			name:      "add to empty counter",
			fields:    fields{},
			increment: 10,
			want:      &counter{0, 10},
		},
		{
			name:      "add to existing counter",
			fields:    fields{7, 8},
			increment: 6,
			want:      &counter{7, 14},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &counter{
				prev:  tt.fields.prev,
				value: tt.fields.value,
			}
			c.Inc(tt.increment)
			assert.Equal(t, tt.want, c)
		})
	}
}

// check gauge update (should we?)
func Test_gauge_Update(t *testing.T) {
	type fields struct {
		gauge    float64
		updateTS time.Time
		reportTS time.Time
	}
	tests := []struct {
		name   string
		fields fields
		gauge  float64
		want   *gauge
	}{
		{
			name: "update empty gauge",
			fields: fields{
				gauge: 0,
			},
			gauge: -0.175,
			want: &gauge{
				gauge: -0.175,
			},
		},
		{
			name: "update existing gauge",
			fields: fields{
				gauge:    12.169,
				updateTS: time.Now().Add(-10 * time.Minute),
			},
			gauge: -0.175,
			want: &gauge{
				gauge: -0.175,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &gauge{
				gauge:    tt.fields.gauge,
				updateTS: tt.fields.updateTS,
				reportTS: tt.fields.reportTS,
			}
			g.Update(tt.gauge)
			assert.Equal(t, tt.want.gauge, g.gauge)
			assert.Greater(t, g.updateTS, tt.fields.updateTS)
		})
	}
}

// check counter delta calculation
func Test_counter_get(t *testing.T) {
	type fields struct {
		prev  int64
		value int64
	}
	tests := []struct {
		name   string
		fields fields
		want   models.Metrics
	}{
		{
			name: "get reported counter",
			fields: fields{
				prev:  12,
				value: 14,
			},
			want: models.Metrics{
				Type:      models.Counter,
				Increment: 2,
				Value:     14,
			},
		},
		{
			name: "get unreported counter",
			fields: fields{
				value: 7,
			},
			want: models.Metrics{
				Type:      models.Counter,
				Increment: 7,
				Value:     7,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &counter{
				prev:  tt.fields.prev,
				value: tt.fields.value,
			}
			if gotM := c.get(); !reflect.DeepEqual(gotM, tt.want) {
				t.Errorf("get() = %v, want %v", gotM, tt.want)
			}
		})
	}
}

// check whether to return gauge or not
func Test_gauge_get(t *testing.T) {
	type fields struct {
		gauge    float64
		updateTS time.Time
		reportTS time.Time
	}
	laterTS := time.Now().Add(-2 * time.Minute)
	earlyTS := time.Now().Add(-3 * time.Minute)
	tests := []struct {
		name   string
		fields fields
		wantM  models.Metrics
		wantOk bool
	}{
		{
			// metric was update, should be returned
			name: "get updated metric",
			fields: fields{
				gauge:    1.001,
				updateTS: laterTS,
				reportTS: earlyTS,
			},
			wantM: models.Metrics{
				Type:      models.Gauge,
				Gauge:     1.001,
				TimeStamp: laterTS,
			},
			wantOk: true,
		},
		{
			// zero metric should not be returned
			name:   "get zero metric",
			fields: fields{},
			wantOk: false,
		},
		{
			// metric that was not updated updateTS < ReportTS should not be returned
			name: "get NOT updated metric",
			fields: fields{
				gauge:    1.001,
				updateTS: earlyTS,
				reportTS: laterTS,
			},
			wantOk: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &gauge{
				gauge:    tt.fields.gauge,
				updateTS: tt.fields.updateTS,
				reportTS: tt.fields.reportTS,
			}
			gotM, gotOk := g.get()
			require.Equal(t, tt.wantOk, gotOk)
			if gotOk {
				assert.Equal(t, tt.wantM, gotM)
			}
		})
	}
}

// check prev updates after reporting
func Test_counter_reported(t *testing.T) {
	type fields struct {
		prev  int64
		value int64
	}
	tests := []struct {
		name          string
		fields        fields
		reportedValue int64
		want          *counter
	}{
		{
			// counter was never reported
			name:          "report unreported counter",
			fields:        fields{0, 16},
			reportedValue: 16,
			want:          &counter{16, 16},
		},
		{
			// counter was already reported
			name:          "report reported counter",
			fields:        fields{19, 21},
			reportedValue: 20,
			want:          &counter{20, 21},
		},
		{
			// this reporting request lasted too long and next report
			//  was already successful. prev shouldn't be updated
			name:          "delayed reported counter",
			fields:        fields{31, 34},
			reportedValue: 29,
			want:          &counter{31, 34},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &counter{
				prev:  tt.fields.prev,
				value: tt.fields.value,
			}
			c.reported(tt.reportedValue)
			assert.Equal(t, tt.want, c)
		})
	}
}

// check reportTS updates after reporting
func Test_gauge_reported(t *testing.T) {
	type fields struct {
		value    float64
		updateTS time.Time
		reportTS time.Time
	}
	prevReportTS := time.Now().Add(-6 * time.Minute)
	currentReportTS := time.Now().Add(-5 * time.Minute)
	nextReportTS := time.Now().Add(-4 * time.Minute)
	updateTS := time.Now().Add(-2 * time.Minute)
	tests := []struct {
		name   string
		fields fields
		want   time.Time
	}{
		{
			name: "gauge was never reported",
			fields: fields{
				value:    10.0,
				updateTS: updateTS,
			},
			want: currentReportTS,
		},
		{
			name: "gauge was already reported",
			fields: fields{
				value:    -0.001,
				updateTS: updateTS,
				reportTS: prevReportTS,
			},
			want: currentReportTS,
		},
		{
			// reporting request lasted too long and the next reporting
			// cycle was already successful. (gouge.reportTS < currentReportTS)
			name: "delay reported gauge",
			fields: fields{
				value:    1.11111,
				updateTS: updateTS,
				reportTS: nextReportTS,
			},
			want: nextReportTS,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &gauge{
				gauge:    tt.fields.value,
				updateTS: tt.fields.updateTS,
				reportTS: tt.fields.reportTS,
			}
			g.reported(currentReportTS)
			assert.Equal(t, tt.want, g.reportTS)
		})
	}
}

type successReporter struct {
	Reporter
}

func (r successReporter) Report(metrics models.Metrics) bool { _ = metrics; return true }

type failedReporter struct {
	Reporter
}

func (r failedReporter) Report(metrics models.Metrics) bool { _ = metrics; return false }

// check update of prev in case of success or failed report
func Test_counter_Report(t *testing.T) {
	type fields struct {
		prev  int64
		value int64
	}
	tests := []struct {
		name     string
		fields   fields
		reporter Reporter
		want     *counter
	}{
		{
			name:     "failed reporting",
			fields:   fields{7, 16},
			reporter: failedReporter{},
			want:     &counter{7, 16},
		},
		{
			name:     "success reporting",
			fields:   fields{7, 16},
			reporter: successReporter{},
			want:     &counter{16, 16},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &counter{
				prev:  tt.fields.prev,
				value: tt.fields.value,
			}
			c.Report("", tt.reporter)
			assert.Equal(t, tt.want, c)
		})
	}
}

// check update of reportTS in case of success or failed report
func Test_gauge_Report(t *testing.T) {
	type fields struct {
		value    float64
		updateTS time.Time
		reportTS time.Time
	}
	updateTS := time.Now().Add(-2 * time.Minute)
	oldReportTS := time.Now().Add(-3 * time.Minute)
	tests := []struct {
		name     string
		fields   fields
		reporter Reporter
		want     time.Time
	}{
		{
			name: "failed reporting",
			fields: fields{
				value:    0.01,
				updateTS: updateTS,
				reportTS: oldReportTS,
			},
			reporter: failedReporter{},
			want:     oldReportTS,
		},
		{
			name: "success reporting",
			fields: fields{
				value:    0.01,
				updateTS: updateTS,
				reportTS: oldReportTS,
			},
			reporter: successReporter{},
			want:     updateTS,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &gauge{
				gauge:    tt.fields.value,
				updateTS: tt.fields.updateTS,
				reportTS: tt.fields.reportTS,
			}
			g.Report("", tt.reporter)
			assert.Equal(t, tt.want, g.reportTS)
		})
	}
}