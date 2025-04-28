package main

import (
	"net/http"
	"github.com/EshkinKot1980/metrics/internal/server/handlers/update"
	"github.com/EshkinKot1980/metrics/internal/server/handlers/value"
	"github.com/EshkinKot1980/metrics/internal/server/storage/memory"
	"github.com/go-chi/chi/v5"
)

func main() {

	storage := memory.New()
	
	router := chi.NewRouter()
	router.Route("/update", func(r chi.Router) {
		r.Post("/{type}/{name}/{value}", update.New(storage))
	})
	router.Get("/value/{type}/{name}", value.New(storage))

	err := http.ListenAndServe(`:8080`, router)
	if err != nil {
		panic(err)
	}
}
