package ipmatcher

import (
	"context"
	"net"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/peer"
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

// contextMutatorUnaryServer interceptor to make context changes
func contextMutatorUnarySrv(mutate func(ctx context.Context) context.Context) func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		return handler(mutate(ctx), req)
	}
}

func dialer(h *mockMetricServer, opts ...grpc.ServerOption) func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(2 << 20)
	server := grpc.NewServer(opts...)
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

func TestIPMatchInterceptor(t *testing.T) {
	//sample request
	req := pb.MetricsBatch{
		Metrics: []*pb.Metric{},
	}
	tests := []struct {
		name     string
		peerIP   string
		subnet   netip.Prefix
		enabled  bool
		wantCode codes.Code
	}{
		{
			name:     "check disabled",
			peerIP:   "127.0.0.1",
			subnet:   netip.MustParsePrefix("10.0.0.0/8"),
			wantCode: codes.OK,
		},
		{
			name:     "peer matches subnet",
			peerIP:   "10.10.20.30",
			enabled:  true,
			subnet:   netip.MustParsePrefix("10.0.0.0/8"),
			wantCode: codes.OK,
		},
		{
			name:     "peer doesn't match subnet",
			peerIP:   "10.10.20.30",
			enabled:  true,
			subnet:   netip.MustParsePrefix("192.168.0.0/16"),
			wantCode: codes.Unauthenticated,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := new(mockMetricServer)
			mutate := func(ctx context.Context) context.Context {
				addr, err := net.ResolveIPAddr("ip", tt.peerIP)
				require.NoError(t, err)
				return peer.NewContext(ctx, &peer.Peer{Addr: addr})
			}
			conn, _ := grpc.Dial(
				"",
				grpc.WithTransportCredentials(insecure.NewCredentials()),
				grpc.WithContextDialer(dialer(ms, grpc.ChainUnaryInterceptor(
					contextMutatorUnarySrv(mutate),
					IPMatchInterceptor(tt.enabled, tt.subnet),
				))),
			)
			defer conn.Close()
			client := pb.NewMetricsClient(conn)
			_, err := client.UpdateMetricsBatch(context.Background(), &req)
			e, ok := status.FromError(err)
			require.Truef(t, ok, "unable to parse grpc status")
			require.Equalf(t, tt.wantCode, e.Code(), "expected code %s got %s", tt.wantCode.String(), e.Code().String())
		})
	}
}
