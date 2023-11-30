package filedump

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/freepaddler/yap-metrics/internal/pkg/models"
)

func TestFileDump_RestoreOnce(t *testing.T) {
	tempFile := fmt.Sprintf("%s/metric_dump", os.TempDir())
	defer os.Remove(tempFile)

	// create dump
	f, err := os.OpenFile(tempFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	require.NoError(t, err)
	c1, _ := models.NewMetric("c1", models.Counter, "10")
	fd := NewFileDump(f)
	fd.Dump([]models.Metrics{c1})
	f.Close()

	// twice restore
	f, err = os.OpenFile(tempFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	require.NoError(t, err)
	defer f.Close()
	fd = NewFileDump(f)
	_, _, err = fd.Restore()
	require.NoError(t, err)
	_, _, err = fd.Restore()
	require.ErrorIs(t, ErrRestoreDenied, err)
}

func TestFileDump_NoRestoreAfterDump(t *testing.T) {
	tempFile := fmt.Sprintf("%s/metric_dump", os.TempDir())
	defer os.Remove(tempFile)

	// create dump
	f, err := os.OpenFile(tempFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	require.NoError(t, err)
	c1, _ := models.NewMetric("c1", models.Counter, "10")
	fd := NewFileDump(f)
	fd.Dump([]models.Metrics{c1})

	_, _, err = fd.Restore()
	require.ErrorIs(t, ErrRestoreDenied, err)
	f.Close()
}

func TestFileDump_DumpRestore(t *testing.T) {
	tempFile := fmt.Sprintf("%s/metric_dump", os.TempDir())
	defer os.Remove(tempFile)

	c1, _ := models.NewMetric("c1", models.Counter, "10")
	c1over, _ := models.NewMetric("c1", models.Counter, "100")
	c2, _ := models.NewMetric("c2", models.Counter, "-10")
	c3, _ := models.NewMetric("c3", models.Counter, "0")
	g1, _ := models.NewMetric("g1", models.Gauge, "-0.117")
	g1over, _ := models.NewMetric("g1", models.Gauge, "-0.117")
	g2, _ := models.NewMetric("g2", models.Gauge, "192.345")
	g3, _ := models.NewMetric("g3", models.Gauge, "0")

	tests := []struct {
		name        string
		runOrder    func() time.Time
		wantRestore []models.Metrics
	}{
		{
			name:        "one dump",
			wantRestore: []models.Metrics{c1, g1, c3},
			runOrder: func() time.Time {
				f, err := os.OpenFile(tempFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
				require.NoError(t, err)
				fd := NewFileDump(f)
				tBefore := time.Now()
				time.Sleep(time.Millisecond)
				fd.Dump([]models.Metrics{c1, g1, c3})
				f.Close()
				return tBefore
			},
		},
		{
			name:        "multiple dumps",
			wantRestore: []models.Metrics{c1over, g1over, c3, g3},
			runOrder: func() time.Time {
				f, err := os.OpenFile(tempFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
				require.NoError(t, err)
				fd := NewFileDump(f)
				fd.Dump([]models.Metrics{c1, g1, c2, g2})
				fd.Dump([]models.Metrics{c3, c1, c2})
				tBefore := time.Now()
				time.Sleep(time.Millisecond)
				fd.Dump([]models.Metrics{c1over, g1over, c3, g3})
				f.Close()
				return tBefore
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer os.Remove(tempFile)
			tBefore := tt.runOrder()
			f, err := os.OpenFile(tempFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
			require.NoError(t, err)
			defer f.Close()

			fd := NewFileDump(f)
			m, ts, err := fd.Restore()
			require.NoError(t, err)
			require.Truef(t, ts.After(tBefore), "Expect %s after %s", ts.Format(time.RFC3339Nano), tBefore.Format(time.RFC3339Nano))
			assert.Equal(t, tt.wantRestore, m)

		})
	}
}
