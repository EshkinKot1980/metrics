package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os/signal"
	"syscall"
	"time"

	"github.com/EshkinKot1980/metrics/internal/agent"
	"github.com/EshkinKot1980/metrics/internal/agent/client"
	"github.com/EshkinKot1980/metrics/internal/agent/monitor"
	"github.com/EshkinKot1980/metrics/internal/agent/storage"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	cfg, err := agent.LoadConfig()
	if err != nil {
		log.Fatal("failed to load config: ", err)
	}

	if cfg.PprofAdrr != "" {
		go http.ListenAndServe(cfg.PprofAdrr, nil)
	}

	s := storage.New()
	m := monitor.New(s)
	am := monitor.NewAdditionalMonitor(s)

	r, err := client.NewReporter(cfg, s)
	if err != nil {
		log.Fatal("failed to init reporter: ", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	go func() {
		for {
			select {
			case <-time.After(cfg.PollInterval):
				m.Poll()
			case <-ctx.Done():
				return
			}

		}
	}()

	go func() {
		for {
			select {
			case <-time.After(cfg.PollInterval):
				am.Poll()
			case <-ctx.Done():
				return
			}

		}
	}()

	for {
		select {
		case <-time.After(cfg.ReportInterval):
			r.Report()
		case <-ctx.Done():
			r.Report()
			return
		}
	}
}
