package main

import (
	"context"
	"fmt"
	"log"
	_ "net/http/pprof"
	"os/signal"
	"syscall"

	"github.com/EshkinKot1980/metrics/internal/server/audit"
	"github.com/EshkinKot1980/metrics/internal/server/config"
	"github.com/EshkinKot1980/metrics/internal/server/http"
	"github.com/EshkinKot1980/metrics/internal/server/logger"
	"github.com/EshkinKot1980/metrics/internal/server/service"
	"github.com/EshkinKot1980/metrics/internal/server/storage"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

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
	httpServer, err := http.NewApp(cfg, service, storage, logger)
	if err != nil {
		return fmt.Errorf("failed to init router: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	return httpServer.Run(ctx)
}
