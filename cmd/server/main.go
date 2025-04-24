package main

import (
	"net/http"
	"github.com/EshkinKot1980/metrics/internal/server/handlers/update"
	"github.com/EshkinKot1980/metrics/internal/server/handlers/value"
	"github.com/EshkinKot1980/metrics/internal/server/storage/memory"
)

func main() {

	storage := memory.New()
	
	mux := http.NewServeMux()
	mux.Handle(`POST /update/{type}/{name}/{value}`, update.New(storage))
	mux.Handle(`GET /value/{type}/{name}`, value.New(storage))
	mux.Handle(`/`, http.NotFoundHandler())

	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
