package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/EshkinKot1980/metrics/internal/common/proto"
	"github.com/EshkinKot1980/metrics/internal/server/audit"
	"github.com/EshkinKot1980/metrics/internal/server/service"
	"github.com/EshkinKot1980/metrics/internal/server/storage"
)

func TestMetricsServer_UpdateMetrics(t *testing.T) {
	type metric struct {
		id    string
		mType pb.Metric_MType
		delta int64
		value float64
	}

	tests := []struct {
		name    string
		metrics []metric
		wantErr error
	}{
		{
			name: "succes",
			metrics: []metric{
				{id: "TestCounter", mType: pb.Metric_COUNTER, delta: 1},
				{id: "TestGauge", mType: pb.Metric_GAUGE, value: 3.14},
			},
		},
		{
			name:    "error_empty_metrics",
			metrics: []metric{},
			wantErr: status.Error(codes.InvalidArgument, "metrics is empty"),
		},
		{
			name: "error_ivalid_type",
			metrics: []metric{
				{id: "TestCounter", mType: pb.Metric_MType(13), delta: 1},
			},
			wantErr: status.Error(codes.InvalidArgument, "invalid metric type"),
		},
	}

	storage := storage.NewMemoryStorage()
	logger := LoggerStub{}
	auditor := AuditorStub{}
	service := service.NewMetricService(storage, logger, auditor)
	server := NewMetricsServer(service)

	buildReq := func(metrics []metric) *pb.UpdateMetricsRequest {
		pbMetrics := make([]*pb.Metric, 0, len(metrics))
		for _, m := range metrics {
			pbm := pb.Metric{}
			pbm.SetId(m.id)
			pbm.SetType(m.mType)
			pbm.SetDelta(m.delta)
			pbm.SetValue(m.value)
			pbMetrics = append(pbMetrics, &pbm)
		}

		return pb.UpdateMetricsRequest_builder{
			Metrics: pbMetrics,
		}.Build()
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := buildReq(test.metrics)
			_, err := server.UpdateMetrics(context.Background(), req)
			assert.ErrorIs(t, err, test.wantErr, "Update metrics error")
		})
	}
}

type LoggerStub struct{}

func (l LoggerStub) Error(message string, err error) {}

type AuditorStub struct{}

func (a AuditorStub) Rise(e audit.Event) {}
