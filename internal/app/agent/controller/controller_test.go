package controller

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/freepaddler/yap-metrics/internal/pkg/models"
	"github.com/freepaddler/yap-metrics/mocks"
)

func TestMetricsController_getMetric(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockMemoryStore(ctrl)

	mc := New(m)

	metrics := models.Metrics{
		Name: "g1",
		Type: models.Gauge,
	}
	m.EXPECT().GetGauge(metrics.Name).Return(-1.1, true).Times(1)
	err := mc.getMetric(&metrics)
	require.NoError(t, err)
	require.Equal(t, *metrics.FValue, -1.1)

	//type fields struct {
	//	store      store.MemoryStore
	//	mu         sync.RWMutex
	//	countersTs map[string]time.Time
	//	gaugesTs   map[string]time.Time
	//}
	//type args struct {
	//	metric *models.Metrics
	//}
	//tests := []struct {
	//	name    string
	//	fields  fields
	//	args    args
	//	wantErr bool
	//}{
	//	// TODO: Add test cases.
	//}
	//for _, tt := range tests {
	//	t.Run(tt.name, func(t *testing.T) {
	//		mc := &MetricsController{
	//			store:      tt.fields.store,
	//			mu:         tt.fields.mu,
	//			countersTs: tt.fields.countersTs,
	//			gaugesTs:   tt.fields.gaugesTs,
	//		}
	//		if err := mc.getMetric(tt.args.metric); (err != nil) != tt.wantErr {
	//			t.Errorf("getMetric() error = %v, wantErr %v", err, tt.wantErr)
	//		}
	//	})
	//}
}
