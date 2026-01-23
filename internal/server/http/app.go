package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/EshkinKot1980/metrics/internal/server/config"
	"github.com/EshkinKot1980/metrics/internal/server/http/handler"
)

type App struct {
	config *config.Config
	router *chi.Mux
}

func NewApp(
	cfg *config.Config,
	srv MetricService,
	p handler.DBPinger,
	l Loger,
) (*App, error) {
	r, err := NewRouter(cfg, srv, p, l)
	if err != nil {
		return nil, fmt.Errorf("failed to init router: %w", err)
	}

	return &App{config: cfg, router: r}, nil
}

func (a *App) Start(ctx context.Context) error {
	srv := &http.Server{Addr: a.config.HTTPaddr, Handler: a.router}
	errChan := make(chan error)

	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return fmt.Errorf("failed to start http server: %w", err)
	case <-time.After(time.Second):
		log.Printf("server listening on %s\n", a.config.HTTPaddr)
	}

	<-ctx.Done()
	log.Println("shutting down http server gracefully")
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		log.Println("http server stopped")
		cancel()
	}()

	return srv.Shutdown(timeoutCtx)
}
