package grpcserver

import (
	"google.golang.org/grpc"

	pb "github.com/freepaddler/yap-metrics/internal/pkg/grpc/proto"
)

type GRCPServer struct {
	Server   *grpc.Server
	address  string
	handlers *GRPCMetricsHandlers
}

func (gs *GRCPServer) GetAddress() string {
	return gs.address
}

func NewGrpcServer(opts ...func(gs *GRCPServer)) *GRCPServer {
	gs := new(GRCPServer)
	for _, o := range opts {
		o(gs)
	}
	gs.Server = grpc.NewServer()
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
