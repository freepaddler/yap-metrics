package sign

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	pb "github.com/freepaddler/yap-metrics/internal/pkg/grpc/proto"
)

type mockMetricServer struct {
	pb.UnimplementedMetricsServer
}

func (ms *mockMetricServer) UpdateMetricsBatch(_ context.Context, _ *pb.MetricsBatch) (*pb.EmptyResponse, error) {
	return &pb.EmptyResponse{}, nil
}

func dialer(opts ...grpc.ServerOption) func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(2 << 20)
	server := grpc.NewServer(opts...)
	pb.RegisterMetricsServer(server, &mockMetricServer{})
	go func() {
		if err := server.Serve(listener); err != nil {
			panic(err)
		}
	}()
	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}

func Test_SignInterceptors(t *testing.T) {
	// test request
	req := pb.MetricsBatch{
		Metrics: []*pb.Metric{
			{
				Id:    "c1",
				Type:  pb.Metric_COUNTER,
				Delta: 10,
			},
		},
	}
	tests := []struct {
		name      string
		clientKey string
		serverKey string
		wantCode  codes.Code
	}{
		{
			name:     "no client key no server key",
			wantCode: codes.OK,
		},
		{
			name:      "same client key and server key",
			clientKey: "somekey",
			serverKey: "somekey",
			wantCode:  codes.OK,
		},
		{
			name:      "different client key and server key",
			clientKey: "somekey1",
			serverKey: "somekey2",
			wantCode:  codes.Unauthenticated,
		},
		{
			name:      "client key no server key",
			clientKey: "somekey",
			wantCode:  codes.OK,
		},
		{
			name:      "no client key server key",
			serverKey: "somekey",
			wantCode:  codes.Unauthenticated,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn, _ := grpc.Dial(
				"",
				grpc.WithTransportCredentials(insecure.NewCredentials()),
				grpc.WithChainUnaryInterceptor(SignGRPCInterceptorClient(tt.clientKey)),
				grpc.WithContextDialer(dialer(grpc.ChainUnaryInterceptor(
					SignCheckGRPCInterceptorServer(tt.serverKey),
				))),
			)
			client := pb.NewMetricsClient(conn)
			_, err := client.UpdateMetricsBatch(context.Background(), &req)
			e, ok := status.FromError(err)
			require.Truef(t, ok, "failed to parse response status")
			require.Equalf(t, tt.wantCode, e.Code(), "expected code %s got %s", tt.wantCode.String(), e.Code().String())
		})
	}
}
