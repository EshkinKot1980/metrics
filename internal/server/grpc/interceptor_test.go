package grpc

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	pb "github.com/EshkinKot1980/metrics/internal/common/proto"
)

func TestFirewall_Filter(t *testing.T) {
	_, trustedNet, err := net.ParseCIDR("172.17.0.0/16")
	require.Nil(t, err, "Parsing trusted network")
	fw := NewFirewall(trustedNet)

	lis := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer(grpc.UnaryInterceptor(fw.Filter))
	pb.RegisterMetricsServer(server, &MockMetricsServer{})
	startTestGrpcServer(t, server, lis)
	defer server.Stop()

	bufDialer := func(ctx context.Context, target string) (net.Conn, error) {
		return lis.Dial()
	}
	conn, err := grpc.NewClient("passthrough://bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.Nil(t, err, "Dialing bufnet")
	defer conn.Close()

	client := pb.NewMetricsClient(conn)
	req := pb.UpdateMetricsRequest_builder{
		Metrics: []*pb.Metric{},
	}.Build()

	tests := []struct {
		name     string
		ip       string
		wantCode codes.Code
	}{
		{
			name:     "success",
			ip:       "172.17.1.2",
			wantCode: codes.OK,
		},
		{
			name:     "error_without_md",
			ip:       "",
			wantCode: codes.PermissionDenied,
		},
		{
			name:     "error_invalid_ip",
			ip:       "InvalidIP",
			wantCode: codes.PermissionDenied,
		},
		{
			name:     "error_untrusted_network",
			ip:       "172.16.1.1",
			wantCode: codes.PermissionDenied,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			kv := make([]string, 0, 2)
			if test.ip != "" {
				kv = append(kv, "x-real-ip", test.ip)
			}

			md := metadata.Pairs(kv...)
			ctx := metadata.NewOutgoingContext(context.Background(), md)
			_, err := client.UpdateMetrics(ctx, req)

			assert.Equal(t, test.wantCode, status.Code(err), "Response status code")
		})
	}

}

func startTestGrpcServer(t *testing.T, s *grpc.Server, l *bufconn.Listener) {
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

type MockMetricsServer struct {
	pb.UnimplementedMetricsServer
}

func (s *MockMetricsServer) UpdateMetrics(
	ctx context.Context,
	req *pb.UpdateMetricsRequest,
) (*pb.UpdateMetricsResponse, error) {
	return &pb.UpdateMetricsResponse{}, nil
}
