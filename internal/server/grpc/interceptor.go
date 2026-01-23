package grpc

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Firewall struct {
	trustedNet *net.IPNet
}

func NewFirewall(trusted *net.IPNet) *Firewall {
	return &Firewall{trustedNet: trusted}
}

func (f *Firewall) Filter(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	if f.trustedNet == nil {
		return handler(ctx, req)
	}

	var clientIP net.IP
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get("x-real-ip")
		if len(values) > 0 {
			clientIP = net.ParseIP(values[0])
		}
	}

	if !f.trustedNet.Contains(clientIP) {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}

	return handler(ctx, req)
}
