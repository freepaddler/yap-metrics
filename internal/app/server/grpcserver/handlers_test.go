package grpcserver

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	pb "github.com/freepaddler/yap-metrics/internal/pkg/grpc/proto"
	"github.com/freepaddler/yap-metrics/internal/pkg/models"
	"github.com/freepaddler/yap-metrics/mocks"
)

func pointer[T any](val T) *T {
	return &val
}

// dialer is a server-client mock
func dialer(h *GRPCMetricsHandlers) func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(2 << 20)
	server := grpc.NewServer()
	pb.RegisterMetricsServer(server, h)
	go func() {
		if err := server.Serve(listener); err != nil {
			panic(err)
		}
	}()
	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}

func TestMetricsServer_UpdateMetricsBatch(t *testing.T) {
	var mockController = gomock.NewController(t)
	defer mockController.Finish()
	m := mocks.NewMockGRPCHandlerStorage(mockController)

	h := NewGRPCHandlers(m)

	conn, err := grpc.Dial("", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(dialer(h)))
	if err != nil {
		require.NoError(t, err)
	}
	defer conn.Close()
	client := pb.NewMetricsClient(conn)

	tests := []struct {
		name        string
		metrics     []*pb.Metric
		wantCalls   int
		wantRequest []models.Metrics
		returnError error
		wantCode    codes.Code
	}{
		{
			name: "one metric",
			metrics: []*pb.Metric{
				{Id: "c1", Type: pb.Metric_COUNTER, Delta: int64(128)},
			},
			wantRequest: []models.Metrics{
				{Name: "c1", Type: models.Counter, IValue: pointer(int64(128))},
			},
			wantCalls: 1,
			wantCode:  codes.OK,
		},
		{
			name: "invalid one",
			metrics: []*pb.Metric{
				{Type: pb.Metric_COUNTER, Delta: int64(128)},
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "valid set",
			metrics: []*pb.Metric{
				{Id: "c1", Type: pb.Metric_COUNTER, Delta: int64(128)},
				{Id: "g1", Type: pb.Metric_GAUGE, Value: -117.09},
			},
			wantRequest: []models.Metrics{
				{Name: "c1", Type: models.Counter, IValue: pointer(int64(128))},
				{Name: "g1", Type: models.Gauge, FValue: pointer(-117.09)},
			},
			wantCalls: 1,
			wantCode:  codes.OK,
		},
		{
			name: "invalid set",
			metrics: []*pb.Metric{
				{Id: "c1", Type: pb.Metric_COUNTER, Delta: int64(128)},
				{Id: "c1", Type: pb.Metric_UNSPECIFIED, Delta: int64(128)},
				{Id: "g1", Type: pb.Metric_GAUGE, Value: -117.09},
			},
			wantCode: codes.InvalidArgument,
		},
		{
			name: "update failed",
			metrics: []*pb.Metric{
				{Id: "c1", Type: pb.Metric_COUNTER, Delta: int64(128)},
			},
			wantRequest: []models.Metrics{
				{Name: "c1", Type: models.Counter, IValue: pointer(int64(128))},
			},
			wantCalls:   1,
			returnError: errors.New("some error"),
			wantCode:    codes.Internal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.EXPECT().
				UpdateMany(gomock.Any()).
				Times(tt.wantCalls).
				DoAndReturn(func(metrics []models.Metrics) error {
					assert.Equal(t, tt.wantRequest, metrics)
					return tt.returnError
				})
			//m.EXPECT().UpdateMany(gomock.Any()).DoAndReturn(func() {}).Times(tt.wantCalls)
			req := pb.MetricsBatch{Metrics: tt.metrics}
			_, err := client.UpdateMetricsBatch(context.TODO(), &req)
			e, ok := status.FromError(err)
			require.Truef(t, ok, "unable to parse grpc response status")
			require.Equalf(t, tt.wantCode, e.Code(), "expected code %s got %s", tt.wantCode.String(), e.Code().String())
		})
	}
}
