package grpcbatchreporter

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	pb "github.com/freepaddler/yap-metrics/internal/pkg/grpc/proto"
	"github.com/freepaddler/yap-metrics/internal/pkg/models"
)

func pointer[T any](val T) *T {
	return &val
}

type mockMetricsServer struct {
	pb.UnimplementedMetricsServer
	// per test method implementation
	funcUpdateMetricsBatch func(ctx context.Context, in *pb.MetricsBatch) (*pb.EmptyResponse, error)
}

func (mock *mockMetricsServer) UpdateMetricsBatch(ctx context.Context, in *pb.MetricsBatch) (*pb.EmptyResponse, error) {
	return mock.funcUpdateMetricsBatch(ctx, in)
}

func dialer(ms *mockMetricsServer) func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(2 << 20)
	server := grpc.NewServer()
	pb.RegisterMetricsServer(server, ms)
	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()
	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}

func TestReporter_Send(t *testing.T) {
	tests := []struct {
		name        string
		grpcRequest bool
		metrics     []models.Metrics
		wantCode    codes.Code
		returnErr   error
	}{
		{
			name:        "empty report",
			grpcRequest: false,
		},
		{
			name:        "ok",
			grpcRequest: true,
			metrics:     []models.Metrics{{Name: "c1", Type: models.Counter, IValue: pointer(int64(10))}},
			wantCode:    codes.OK,
		},
		{
			name:        "error",
			grpcRequest: true,
			metrics:     []models.Metrics{{Name: "c1", Type: models.Counter, IValue: pointer(int64(10))}},
			returnErr:   status.Error(codes.InvalidArgument, "some"),
			wantCode:    codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockMetricsServer{}
			ctx := context.Background()
			conn, err := grpc.DialContext(
				ctx,
				"",
				grpc.WithTransportCredentials(insecure.NewCredentials()),
				grpc.WithContextDialer(dialer(mock)),
			)
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			r, err := New(
				WithTimeout(time.Second),
				WithAddress("240.0.0.0:65535"),
			)
			require.NoError(t, err)
			r.conn = conn
			mock.funcUpdateMetricsBatch = func(_ context.Context, in *pb.MetricsBatch) (*pb.EmptyResponse, error) {
				require.True(t, tt.grpcRequest)
				return &pb.EmptyResponse{}, tt.returnErr
			}
			if tt.grpcRequest {
				err = r.Send(tt.metrics)
				e, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, tt.wantCode, e.Code())
			}

		})
	}
}
