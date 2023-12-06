package grpcserver

import (
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/encoding/gzip"

	pb "github.com/freepaddler/yap-metrics/internal/pkg/grpc/proto"
	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
)

type GRCPServer struct {
	Server       *grpc.Server
	address      string
	handlers     *GRPCMetricsHandlers
	interceptors []grpc.UnaryServerInterceptor
}

func (gs *GRCPServer) GetAddress() string {
	return gs.address
}

func NewGrpcServer(opts ...func(gs *GRCPServer)) *GRCPServer {
	gs := new(GRCPServer)
	// setup logging
	logOpts := []logging.Option{
		logging.WithLogOnEvents(logging.FinishCall),
		logging.WithFieldsFromContext(logger.GRPCContextFields),
	}
	gs.interceptors = append(
		gs.interceptors,
		logging.UnaryServerInterceptor(logger.GRPCInterceptorLogger(*logger.Log()), logOpts...),
	)
	for _, o := range opts {
		o(gs)
	}
	gs.Server = grpc.NewServer(grpc.ChainUnaryInterceptor(gs.interceptors...))
	pb.RegisterMetricsServer(gs.Server, gs.handlers)
	return gs
}

func WithAddress(s string) func(gs *GRCPServer) {
	return func(gs *GRCPServer) {
		gs.address = s
	}
}

func WithHandlers(h *GRPCMetricsHandlers) func(gs *GRCPServer) {
	return func(gs *GRCPServer) {
		gs.handlers = h

	}
}

func WithInterceptors(interceptors ...grpc.UnaryServerInterceptor) func(server *GRCPServer) {
	return func(gs *GRCPServer) {
		// order makes sense
		gs.interceptors = append(gs.interceptors, interceptors...)
	}
}
