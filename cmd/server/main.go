package main

import (
	"flag"
	"net/http"
	"github.com/go-chi/chi/v5"
	
	"github.com/EshkinKot1980/metrics/internal/server/handlers/update"
	"github.com/EshkinKot1980/metrics/internal/server/handlers/value"
	"github.com/EshkinKot1980/metrics/internal/server/storage/memory"
)

func main() {
	addr := flag.String("a", "localhost:8080", "address to serve in format host:port")
	flag.Parse()

	storage := memory.New()	
	router := chi.NewRouter()
	router.Route("/update", func(r chi.Router) {
		r.Post("/{type}/{name}/{value}", update.New(storage))
	})
	router.Get("/value/{type}/{name}", value.New(storage))

	err := http.ListenAndServe(*addr, router)
	if err != nil {
		panic(err)
	}
}


