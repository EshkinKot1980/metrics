// Модуль http реализует серверную часть приложения сбора метрик.
package http

import (
	"crypto/rsa"
	"fmt"
	"net"

	"github.com/go-chi/chi/v5"
	chiMW "github.com/go-chi/chi/v5/middleware"

	"github.com/EshkinKot1980/metrics/internal/common/utils"
	"github.com/EshkinKot1980/metrics/internal/server/config"
	"github.com/EshkinKot1980/metrics/internal/server/http/handler"
	"github.com/EshkinKot1980/metrics/internal/server/http/middleware"
)

// Cервис для работы с метриками, объединяет сервисы из пакета handler.
type MetricService interface {
	handler.RetrieveService
	handler.UpdateService
}

// Логгер, объединяет логгер из пакета handler с логгером из пакета middleware.
type Loger interface {
	handler.Logger
	middleware.HTTPLogWriter
}

// Инициализация роутера.
func NewRouter(
	cfg *config.Config,
	srv MetricService,
	p handler.DBPinger,
	l Loger,
) (*chi.Mux, error) {
	var (
		privateKey *rsa.PrivateKey
		trustedNet *net.IPNet
		err        error
	)

	if cfg.PrivateKey != "" {
		privateKey, err = utils.LoadPrivateKey(cfg.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to load private key: %w", err)
		}
	}

	if cfg.TrustedSubnet != "" {
		_, trustedNet, err = net.ParseCIDR(cfg.TrustedSubnet)
		if err != nil {
			return nil, fmt.Errorf("failed to parse trusted subnet: %w", err)
		}
	}

	mwLogger := middleware.NewHTTPLogger(l)
	mwHashHeader := middleware.NewHashHeader(cfg.SecretKey)
	mwRSA := middleware.NewRSA(privateKey)
	mwFirewall := middleware.NewFirewall(trustedNet)
	updater := handler.NewUpdateHandler(srv, l)
	retriever := handler.NewRetrieveHandler(srv, l)
	pinger := handler.NewPingHandler(p)

	router := chi.NewRouter()
	router.Use(chiMW.RealIP)
	router.Use(mwLogger.Log)
	router.Use(middleware.GzipWrapper)
	router.Use(mwHashHeader.Sign)

	// В прродакшине тут будет ограничение по IP
	router.Mount("/debug", chiMW.Profiler())

	router.Route("/update", func(r chi.Router) {
		r.Post("/{type}/{name}/{value}", updater.UpdateFromPath)
		r.Post("/", updater.Update)
	})
	router.Route("/updates", func(r chi.Router) {
		r.Use(mwFirewall.Filter)
		r.Use(mwRSA.Decrypt)
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

	return router, nil
}
