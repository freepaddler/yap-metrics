package grpcbatchreporter

import (
	"context"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/encoding/gzip"

	pb "github.com/freepaddler/yap-metrics/internal/pkg/grpc/proto"
	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
	"github.com/freepaddler/yap-metrics/internal/pkg/models"
)

type Reporter struct {
	timeout      time.Duration
	address      string
	conn         *grpc.ClientConn
	interceptors []grpc.UnaryClientInterceptor
	encoder      encoding.Codec
}

func (r *Reporter) Close() error {
	return r.conn.Close()
}

func New(opts ...func(r *Reporter)) (*Reporter, error) {
	reporter := &Reporter{timeout: 5 * time.Second}
	for _, opt := range opts {
		opt(reporter)
	}
	if reporter.encoder != nil {
		encoding.RegisterCodec(reporter.encoder)
	}
	// setup logging
	logOpts := []logging.Option{
		logging.WithLogOnEvents(logging.FinishCall),
	}
	reporter.interceptors = append(
		reporter.interceptors,
		logging.UnaryClientInterceptor(logger.GRPCInterceptorLogger(*logger.Log()), logOpts...),
	)
	var err error
	reporter.conn, err = grpc.Dial(
		reporter.address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.CallContentSubtype("rsakeypair")),
		grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)),
		grpc.WithChainUnaryInterceptor(reporter.interceptors...),
	)
	if err != nil {
		logger.Log().Error().Err(err).Msgf("unable to setup grpc connection to %s", reporter.address)
		return nil, err
	}
	return reporter, nil
}

func WithAddress(a string) func(*Reporter) {
	return func(r *Reporter) {
		r.address = a
	}
}

func WithTimeout(d time.Duration) func(*Reporter) {
	return func(r *Reporter) {
		r.timeout = d
	}
}

func WithInterceptors(interceptors ...grpc.UnaryClientInterceptor) func(*Reporter) {
	return func(r *Reporter) {
		// order makes sense
		r.interceptors = append(r.interceptors, interceptors...)
	}
}

func WithEncoder(codec encoding.Codec) func(reporter *Reporter) {
	return func(r *Reporter) {
		r.encoder = codec
	}
}

func (r *Reporter) Send(m []models.Metrics) (err error) {
	log := logger.Log().With().Str("module", "grpcBatchReporter").Logger()
	if len(m) == 0 {
		log.Info().Msg("skip sending: empty report")
		return
	}
	req := &pb.MetricsBatch{Metrics: make([]*pb.Metric, 0, len(m))}
	for _, v := range m {
		var metric pb.Metric
		if err := metric.FromMetrics(v); err != nil {
			log.Error().Err(err).Msgf("unable to serialize metric %+v", v)
			return err
		}
		req.Metrics = append(req.Metrics, &metric)
	}
	c := pb.NewMetricsClient(r.conn)
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	log.Debug().Msgf("sending %d metrics", len(req.Metrics))
	_, err = c.UpdateMetricsBatch(ctx, req)
	if err != nil {
		log.Error().Err(err).Msgf("grpc request failed: %s", err.Error())
		return err
	}
	return nil
}
