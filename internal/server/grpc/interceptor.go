package grpc

import (
	"context"
	"net"
	"strings"

	"github.com/EshkinKot1980/metrics/internal/server/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
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
		var ip string
		p, ok := peer.FromContext(ctx)
		if ok {
			addr := strings.Split(p.Addr.String(), ":")
			ip = addr[0]
		}
		ctx = context.WithValue(ctx, service.KeyClientIP, ip)

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

	ctx = context.WithValue(ctx, service.KeyClientIP, clientIP.String())

	return handler(ctx, req)
}
