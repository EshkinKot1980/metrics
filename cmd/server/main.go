package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"github.com/EshkinKot1980/metrics/internal/server/handlers/update"
	"github.com/EshkinKot1980/metrics/internal/server/handlers/value"
	"github.com/EshkinKot1980/metrics/internal/server/storage/memory"
)

func main() {
	//TODO: сделать нормальный конфиг c настройками сервера
	addr := loadAddr()
	storage := memory.New()
	router := chi.NewRouter()

	router.Route("/update", func(r chi.Router) {
		r.Post("/{type}/{name}/{value}", update.New(storage))
	})
	router.Get("/value/{type}/{name}", value.New(storage))

	err := http.ListenAndServe(addr, router)
	if err != nil {
		panic(err)
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
