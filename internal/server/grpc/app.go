package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"

	pb "github.com/EshkinKot1980/metrics/internal/common/proto"
	"github.com/EshkinKot1980/metrics/internal/server/config"
)

type App struct {
	cfg     *config.Config
	srv     MetricsService
	stopped chan struct{}
}

func NewApp(c *config.Config, s MetricsService) *App {
	return &App{
		cfg:     c,
		srv:     s,
		stopped: make(chan struct{}),
	}
}

func (a *App) Start(ctx context.Context) error {
	var trustedNet *net.IPNet
	var err error

	if a.cfg.TrustedSubnet != "" {
		_, trustedNet, err = net.ParseCIDR(a.cfg.TrustedSubnet)
		if err != nil {
			return fmt.Errorf("failed to parse trusted subnet: %w", err)
		}
	}

	firewal := NewFirewall(trustedNet)
	server := grpc.NewServer(grpc.UnaryInterceptor(firewal.Filter))
	ms := NewMetricsServer(a.srv)
	pb.RegisterMetricsServer(server, ms)
	errChan := make(chan error)

	listen, err := net.Listen("tcp", a.cfg.GRPCaddr)
	if err != nil {
		return fmt.Errorf("failed to bind address and port: %w", err)
	}

	go func() {
		if err := server.Serve(listen); err != nil {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return fmt.Errorf("failed start grpc server: %w", err)
	case <-time.After(time.Second):
		log.Printf("grpc server listening on %s\n", a.cfg.GRPCaddr)
	}

	go func() {
		<-ctx.Done()
		log.Println("shutting down grpc server gracefully")
		server.GracefulStop()
		log.Println("grpc server stopped")
		close(a.stopped)
	}()

	return nil
}

func (a *App) Stop() {
	<-a.stopped
}
