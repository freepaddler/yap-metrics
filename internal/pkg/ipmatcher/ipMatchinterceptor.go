package ipmatcher

import (
	"context"
	"net/netip"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
)

func IPMatchInterceptor(enabled bool, subnet netip.Prefix) func(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if !enabled {
			return handler(ctx, req)
		}
		if p, ok := peer.FromContext(ctx); ok {
			ip, err := netip.ParseAddr(strings.Split(p.Addr.String(), ":")[0])
			if err != nil {
				logger.Log().Warn().Err(err).Msgf("failed to parse grpc peer address %s", p.Addr.String())
				return nil, status.Errorf(codes.Unauthenticated, "unable parse peer address")
			}
			if subnet.Contains(ip) {
				return handler(ctx, req)
			}
			logger.Log().Warn().Msgf("unauthorized access from %s", ip.String())
			return nil, status.Errorf(codes.Unauthenticated, "access denied")
		} else {
			logger.Log().Warn().Msgf("failed to get grpc peer address")
			return nil, status.Errorf(codes.Unauthenticated, "unable to get peer address")
		}
	}
}
