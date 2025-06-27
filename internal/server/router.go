package server

import (
	"database/sql"
	"github.com/go-chi/chi/v5"

	"github.com/EshkinKot1980/metrics/internal/server/handlers/info"
	"github.com/EshkinKot1980/metrics/internal/server/handlers/ping"
	"github.com/EshkinKot1980/metrics/internal/server/handlers/retrieve"
	"github.com/EshkinKot1980/metrics/internal/server/handlers/update"
	"github.com/EshkinKot1980/metrics/internal/server/handlers/updates"
	"github.com/EshkinKot1980/metrics/internal/server/logger"
	"github.com/EshkinKot1980/metrics/internal/server/middleware"
	"github.com/EshkinKot1980/metrics/internal/server/storage"
)

func NewRouter(config *Config, storage storage.Storage, logger *logger.Logger, db *sql.DB) *chi.Mux {
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
