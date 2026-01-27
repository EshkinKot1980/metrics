package grpc

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/EshkinKot1980/metrics/internal/agent"
	"github.com/EshkinKot1980/metrics/internal/agent/storage"
	pb "github.com/EshkinKot1980/metrics/internal/common/proto"
)

func testRequest(ctx context.Context, req *pb.UpdateMetricsRequest, wantedIP string) func(t *testing.T) {
	return func(t *testing.T) {
		var clientIP net.IP
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			values := md.Get("x-real-ip")
			if len(values) > 0 {
				clientIP = net.ParseIP(values[0])
			}
		}

		assert.Equal(t, wantedIP, clientIP.String(), "Request metadata x-real-ip header")

		metrics := req.GetMetrics()
		assert.Equal(t, 4, len(metrics), "Request metrics count")

		ids := []string{"TestCounter", "Visitors", "ConstE", "TTL"}
		types := []pb.Metric_MType{pb.Metric_COUNTER, pb.Metric_GAUGE}

		for _, m := range metrics {
			assert.Contains(t, ids, m.GetId(), "Request metrics IDs")
			assert.Contains(t, types, m.GetType(), "Request metrics types")
		}
	}
}

func TestClient_Report(t *testing.T) {
	agentIP := "172.17.0.2"
	storage := storage.New()
	testInitStorage(storage)

	lis, err := net.Listen("tcp", "localhost:0")
	require.Nil(t, err, "Listen unused localhost tcp port")

	server := grpc.NewServer()
	pb.RegisterMetricsServer(server, newMockMetricsServer(t, agentIP))
	startTestGrpcServer(t, server, lis)
	defer server.Stop()

	c := NewClient(storage, lis.Addr().String(), agentIP)
	c.Report()
}

type mockMetricsServer struct {
	pb.UnimplementedMetricsServer
	t  *testing.T
	ip string
}

func newMockMetricsServer(t *testing.T, agentIP string) *mockMetricsServer {
	return &mockMetricsServer{t: t, ip: agentIP}
}

func (s *mockMetricsServer) UpdateMetrics(
	ctx context.Context,
	req *pb.UpdateMetricsRequest,
) (*pb.UpdateMetricsResponse, error) {
	s.t.Run("rquest", testRequest(ctx, req, s.ip))
	return &pb.UpdateMetricsResponse{}, nil
}

func startTestGrpcServer(t *testing.T, s *grpc.Server, l net.Listener) {
	errChan := make(chan error)
	go func() {
		err := s.Serve(l)
		if err != nil && err != grpc.ErrServerStopped {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		t.Fatalf("Failed to serve: %v", err)
	case <-time.After(10 * time.Millisecond):
	}
}

func testInitStorage(s *storage.MemoryStorage) {
	s.Put(
		[]agent.Counter{
			{Name: "TestCounter", Value: 13},
			{Name: "Visitors", Value: 256},
		},
		[]agent.Gauge{
			{Name: "ConstE", Value: 2.71828},
			{Name: "TTL", Value: 3.14e50},
		},
	)
}
