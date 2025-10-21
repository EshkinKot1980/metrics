package main

import (
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/EshkinKot1980/metrics/internal/agent"
	"github.com/EshkinKot1980/metrics/internal/agent/client"
	oldAPIclient "github.com/EshkinKot1980/metrics/internal/agent/client/compatible"
	"github.com/EshkinKot1980/metrics/internal/agent/monitor"
	"github.com/EshkinKot1980/metrics/internal/agent/storage"
)

type reporter interface {
	Report()
}

func main() {
	cfg := agent.MustLoadConfig()

	if cfg.PprofAdrr != "" {
		go http.ListenAndServe(cfg.PprofAdrr, nil)
	}

	s := storage.New()
	m := monitor.New(s)
	am := monitor.NewAdditionalMonitor(s)

	var r reporter
	if cfg.BatchReport {
		r = client.New(s, cfg.BaseURL, cfg.SecretKey)
	} else {
		r = oldAPIclient.New(s, cfg.BaseURL, cfg.RateLimit)
	}

	pollInterval := time.Duration(cfg.PollInterval) * time.Second

	go func() {
		for {
			<-time.After(pollInterval)
			m.Poll()
		}
	}()

	go func() {
		for {
			<-time.After(pollInterval)
			am.Poll()
		}
	}()

	interval := time.Duration(cfg.ReportInterval) * time.Second
	for {
		<-time.After(interval)
		r.Report()
	}
}
