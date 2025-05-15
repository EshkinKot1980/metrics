package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"github.com/EshkinKot1980/metrics/internal/server"
	"github.com/EshkinKot1980/metrics/internal/server/handlers/retrieve"
	"github.com/EshkinKot1980/metrics/internal/server/handlers/update"
	"github.com/EshkinKot1980/metrics/internal/server/middleware"
	"github.com/EshkinKot1980/metrics/internal/server/storage/memory"
)

func main() {
	//TODO: сделать нормальный конфиг c настройками сервера
	addr := loadAddr()
	logger := server.MustSetupLogger()
	defer logger.Sync()

	storage := memory.New()
	mvLogger := middleware.NewHTTPLogger(logger)
	updaterHandler := update.New(storage, logger)
	retrieverHandler := retrieve.New(storage, logger)
	router := chi.NewRouter()

	router.Use(mvLogger.Log)
	router.Route("/update/{type}/{name}/{value}", func(r chi.Router) {
		r.Use(middleware.ValidateMetric)
		r.Post("/", updaterHandler.Update)
	})
	router.Get("/value/{type}/{name}", retrieverHandler.Retrieve)

	err := http.ListenAndServe(addr, router)
	if err != nil {
		log.Fatal(err)
	}
}

func loadAddr() string {
	var addr string
	flag.StringVar(&addr, "a", "localhost:8080", "address to serve")

	flag.Parse()

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		addr = envAddr
	}

	return addr
}
