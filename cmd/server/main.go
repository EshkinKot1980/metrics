package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/EshkinKot1980/metrics/internal/server/handlers/retrieve"
	"github.com/EshkinKot1980/metrics/internal/server/handlers/update"
	"github.com/EshkinKot1980/metrics/internal/server/middleware"
	"github.com/EshkinKot1980/metrics/internal/server/storage/memory"
	"github.com/go-chi/chi/v5"
)

func main() {
	//TODO: сделать нормальный конфиг c настройками сервера
	addr := loadAddr()
	storage := memory.New()
	updaterHandler := update.New(storage)
	retrieverHandler := retrieve.New(storage)
	router := chi.NewRouter()

	router.Route("/update", func(r chi.Router) {
		r.Post("/{type}/{name}/{value}", middleware.ValidateMetric(http.HandlerFunc(updaterHandler.Update)))
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
