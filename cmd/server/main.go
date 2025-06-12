package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/EshkinKot1980/metrics/internal/server"
	"github.com/EshkinKot1980/metrics/internal/server/handlers/info"
	"github.com/EshkinKot1980/metrics/internal/server/handlers/ping"
	"github.com/EshkinKot1980/metrics/internal/server/handlers/retrieve"
	"github.com/EshkinKot1980/metrics/internal/server/handlers/update"
	"github.com/EshkinKot1980/metrics/internal/server/handlers/updates"
	"github.com/EshkinKot1980/metrics/internal/server/middleware"
	"github.com/EshkinKot1980/metrics/internal/server/storage"
	"github.com/EshkinKot1980/metrics/internal/server/storage/file"
	"github.com/EshkinKot1980/metrics/internal/server/storage/pg"
)

func main() {
	config := server.MustLoadConfig()
	logger := server.MustSetupLogger()
	defer logger.Sync()

	storage, db := mustSetupStorage(config, logger)
	defer func() {
		storage.Halt()
		db.Close()
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	router := setupRouter(config, storage, logger, db)
	runServer(ctx, config.ServerAddr, router)
}

func runServer(ctx context.Context, addr string, router *chi.Mux) {
	srv := &http.Server{Addr: addr, Handler: router}

	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
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

func setupRouter(config *server.Config, storage storage.Storage, logger *server.Logger, db *sql.DB) *chi.Mux {
	mwLogger := middleware.NewHTTPLogger(logger)
	mwHashHeader := middleware.NewHashHeader(config.SecretKey)
	updaterHandler := update.New(storage, logger)
	updaterJSONHandler := update.NewJSONHandler(storage, logger)
	updaterBatchHandler := updates.New(storage, logger)
	retrieverHandler := retrieve.New(storage, logger)
	retrieverJSONHandler := retrieve.NewJSONHandler(storage, logger)
	pingHandler := ping.New(db)

	router := chi.NewRouter()
	router.Use(mwLogger.Log)
	router.Use(middleware.GzipWrapper)
	router.Use(mwHashHeader.Sign)

	router.Route("/update", func(r chi.Router) {
		r.Post("/{type}/{name}/{value}", updaterHandler.Update)
		r.Post("/", updaterJSONHandler.Update)
	})
	router.Route("/updates", func(r chi.Router) {
		r.Use(mwHashHeader.Validate)
		r.Post("/", updaterBatchHandler.Update)
	})
	router.Route("/value", func(r chi.Router) {
		r.Get("/{type}/{name}", retrieverHandler.Retrieve)
		r.Post("/", retrieverJSONHandler.Retrieve)
	})
	router.Route("/ping", func(r chi.Router) {
		r.Get("/", pingHandler.Ping)
	})
	router.Get("/", info.InfoPage)

	return router
}

func mustSetupStorage(config *server.Config, logger *server.Logger) (storage.Storage, *sql.DB) {
	db, err := sql.Open("pgx", config.DatabaseDSN)
	if err != nil {
		log.Fatal(err)
	}

	if config.DatabaseDSN != "" {
		if err := db.Ping(); err != nil {
			log.Fatal(err)
		}

		storage, err := pg.New(db)
		if err != nil {
			log.Fatal(err)
		}

		return storage, db
	}

	storage, err := file.New(config.FileCfg, logger)
	if err != nil {
		log.Fatal(err)
	}

	return storage, db
}
