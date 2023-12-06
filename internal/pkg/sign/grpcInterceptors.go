package sign

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
)

const (
	HeaderSignGRPC = "HashSHA256"
)

func SignGRPCInterceptorClient(key string) func(
	ctx context.Context,
	method string,
	req interface{},
	reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	return func(ctx context.Context, method string, req interface{}, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		nCtx := ctx
		if key != "" {
			logger.Log().Debug().Msgf("request to sign: %+v", req)
			hash := Get([]byte(getRequestString(req, method)), key)
			nCtx = metadata.AppendToOutgoingContext(ctx, HeaderSignGRPC, hash)
		}
		return invoker(nCtx, method, req, reply, cc, opts...)
	}
}

func SignCheckGRPCInterceptorServer(key string) func(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if key != "" {
			logger.Log().Debug().Msgf("validating signature of request: %+v", req)
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return nil, status.Errorf(codes.Unauthenticated, "unable to parse request metadata")
			}
			signature := md.Get(HeaderSignGRPC)
			if len(signature) < 1 || len(signature[0]) < 2 {
				return nil, status.Errorf(codes.Unauthenticated, "unable to get signature header")
			}
			hash := Get([]byte(getRequestString(req, info.FullMethod)), key)
			if hash != signature[0] {
				return nil, status.Errorf(codes.Unauthenticated, "invalid signature")
			}
		}
		return handler(ctx, req)
	}
}

// formats string from request
func getRequestString(req interface{}, method string) string {
	logger.Log().Debug().Msgf("request method is: %s", method)
	return fmt.Sprintf("request=%v;method=%s", req, method)
}
