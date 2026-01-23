package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/EshkinKot1980/metrics/internal/common/models"
	pb "github.com/EshkinKot1980/metrics/internal/common/proto"
)

// Сервис сохранения метрик.
type MetricsService interface {
	// Сохраняет множество метрик.
	PutList(ctx context.Context, metrics []models.Metrics) error
}

// GRPC серврер для работы с метриками
type MetricsServer struct {
	pb.UnimplementedMetricsServer
	service MetricsService
}

func NewMetricsServer(s MetricsService) *MetricsServer {
	return &MetricsServer{service: s}
}

// Обновляет множество метрик из массива в запросе
func (s *MetricsServer) UpdateMetrics(
	ctx context.Context,
	req *pb.UpdateMetricsRequest,
) (*pb.UpdateMetricsResponse, error) {
	reqMetrics := req.GetMetrics()
	if len(reqMetrics) == 0 {
		return nil, status.Error(codes.InvalidArgument, "metrics is empty")
	}

	metrics := make([]models.Metrics, 0, len(reqMetrics))
	for _, rm := range reqMetrics {
		var m models.Metrics

		switch rm.GetType() {
		case pb.Metric_GAUGE:
			v := rm.GetValue()
			m = models.Metrics{ID: rm.GetId(), MType: models.TypeGauge, Value: &v}
			metrics = append(metrics, m)
		case pb.Metric_COUNTER:
			d := rm.GetDelta()
			m = models.Metrics{ID: rm.GetId(), MType: models.TypeCounter, Delta: &d}
			metrics = append(metrics, m)
		default:
			m = models.Metrics{ID: rm.GetId(), MType: rm.GetType().String()}
		}

		if err := m.Validate(); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	err := s.service.PutList(ctx, metrics)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &pb.UpdateMetricsResponse{}, nil
}
