package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/EshkinKot1980/metrics/internal/server"
	"github.com/EshkinKot1980/metrics/internal/server/logger"
	"github.com/EshkinKot1980/metrics/internal/server/storage"
	"github.com/EshkinKot1980/metrics/internal/server/storage/file"
	"github.com/EshkinKot1980/metrics/internal/server/storage/pg"
)

func main() {
	config := server.MustLoadConfig()

	logger, err := logger.New()
	if err != nil {
		log.Fatal("failed to init logger: ", err)
	}
	defer logger.Sync()

	db, err := sql.Open("pgx", config.DatabaseDSN)
	if err != nil {
		log.Fatal("failed to open database: ", err)
	}
	defer db.Close()

	storage, err := makeStorage(config, db, logger)
	if err != nil {
		db.Close()
		log.Fatal(err)
	}
	defer storage.Halt()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	router := server.NewRouter(config, storage, logger, db)
	runServer(ctx, config.ServerAddr, router, db)
}

func runServer(ctx context.Context, addr string, router http.Handler, db *sql.DB) {
	srv := &http.Server{Addr: addr, Handler: router}

	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			db.Close()
			log.Fatal(err)
		}
	}()

	log.Printf("server listening on %s\n", addr)

	<-ctx.Done()
	log.Println("shutting down server gracefully")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Println(err)
	}

	<-shutdownCtx.Done()
	log.Println("server stopped")
}

func makeStorage(config *server.Config, db *sql.DB, logger *logger.Logger) (storage.Storage, error) {
	if config.DatabaseDSN != "" {
		if err := db.Ping(); err != nil {
			return nil, fmt.Errorf("database is not reachable: %w", err)
		}
		storage, err := pg.New(db)
		if err != nil {
			return nil, fmt.Errorf("failed to create storage: %w", err)
		}
		return storage, nil
	}

	storage, err := file.New(config.FileCfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage: %w", err)
	}

	return storage, nil
}
