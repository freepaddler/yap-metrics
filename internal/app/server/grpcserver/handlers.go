package grpcserver

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/freepaddler/yap-metrics/internal/pkg/grpc/proto"
	"github.com/freepaddler/yap-metrics/internal/pkg/logger"
	"github.com/freepaddler/yap-metrics/internal/pkg/models"
)

//go:generate mockgen -source $GOFILE -package=mocks -destination ../../../../mocks/GRPCHandlerStorage_mock.go

type GRPCHandlerStorage interface {
	UpdateMany(metrics []models.Metrics) error
	Ping() error
}

type GRPCMetricsHandlers struct {
	pb.UnimplementedMetricsServer
	storage GRPCHandlerStorage
}

func NewGRPCHandlers(storage GRPCHandlerStorage) *GRPCMetricsHandlers {
	return &GRPCMetricsHandlers{storage: storage}
}

func (s GRPCMetricsHandlers) UpdateMetricsBatch(_ context.Context, in *pb.MetricsBatch) (*pb.EmptyResponse, error) {
	logger.Log().Debug().Msg("GRPC UpdateMetricsBatch: Request received")
	var response pb.EmptyResponse
	metrics := make([]models.Metrics, 0, len(in.Metrics))
	for _, v := range in.Metrics {
		m, err := v.ToMetrics()
		if err != nil {
			logger.Log().Warn().Err(err).Msgf("unable to parse metric %v", v)
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		metrics = append(metrics, m)
	}
	logger.Log().Info().Msgf("GRPC UpdateMetricsBatch: updating %d metrics", len(metrics))
	err := s.storage.UpdateMany(metrics)
	if err != nil {
		logger.Log().Warn().Err(err).Msg("unable to update metrics")
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &response, nil
}
