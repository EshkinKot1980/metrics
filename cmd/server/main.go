package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os/signal"
	"syscall"
	"time"

	"github.com/EshkinKot1980/metrics/internal/server"
	"github.com/EshkinKot1980/metrics/internal/server/audit"
	"github.com/EshkinKot1980/metrics/internal/server/config"
	"github.com/EshkinKot1980/metrics/internal/server/logger"
	"github.com/EshkinKot1980/metrics/internal/server/service"
	"github.com/EshkinKot1980/metrics/internal/server/storage"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	logger, err := logger.New()
	if err != nil {
		return fmt.Errorf("failed to init logger: %w", err)
	}
	defer logger.Sync()

	storage, err := storage.New(cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to init storage: %w", err)
	}
	defer storage.Halt()

	auditor := audit.NewAuditor(cfg, logger)
	defer auditor.Halt()

	service := service.NewMetricService(storage, logger, auditor)
	router := server.NewRouter(cfg, service, storage, logger)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	return runServer(ctx, cfg.ServerAddr, router)
}

func runServer(ctx context.Context, addr string, router http.Handler) error {
	srv := &http.Server{Addr: addr, Handler: router}
	errChan := make(chan error)

	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-time.After(time.Second):
		log.Printf("server listening on %s\n", addr)
	}

	<-ctx.Done()
	log.Println("shutting down http server gracefully")
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		log.Println("http server stopped")
		cancel()
	}()

	return srv.Shutdown(timeoutCtx)
}
