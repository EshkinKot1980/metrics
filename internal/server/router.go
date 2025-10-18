package server

import (
	"github.com/go-chi/chi/v5"
	chiMW "github.com/go-chi/chi/v5/middleware"

	"github.com/EshkinKot1980/metrics/internal/server/config"
	"github.com/EshkinKot1980/metrics/internal/server/handler"
	"github.com/EshkinKot1980/metrics/internal/server/middleware"
)

type MetricService interface {
	handler.RetrieveService
	handler.UpdateService
}

type Loger interface {
	handler.Logger
	middleware.HTTPLogWriter
}

func NewRouter(
	cfg *config.Config,
	srv MetricService,
	p handler.DBPinger,
	l Loger,
) *chi.Mux {
	mwLogger := middleware.NewHTTPLogger(l)
	mwHashHeader := middleware.NewHashHeader(cfg.SecretKey)
	updater := handler.NewUpdateHandler(srv, l)
	retriever := handler.NewRetrieveHandler(srv, l)
	pinger := handler.NewPingHandler(p)

	router := chi.NewRouter()
	router.Use(chiMW.RealIP)
	router.Use(mwLogger.Log)
	router.Use(middleware.GzipWrapper)
	router.Use(mwHashHeader.Sign)

	router.Route("/update", func(r chi.Router) {
		r.Post("/{type}/{name}/{value}", updater.UpdateFromPath)
		r.Post("/", updater.Update)
	})
	router.Route("/updates", func(r chi.Router) {
		r.Use(mwHashHeader.Validate)
		r.Post("/", updater.UpdateList)
	})
	router.Route("/value", func(r chi.Router) {
		r.Get("/{type}/{name}", retriever.GetByPath)
		r.Post("/", retriever.GetJSON)
	})
	router.Route("/ping", func(r chi.Router) {
		r.Get("/", pinger.Ping)
	})
	router.Get("/", handler.InfoPage)

	return router
}
