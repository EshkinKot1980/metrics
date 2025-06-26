package main

import (
	"time"

	"github.com/EshkinKot1980/metrics/internal/agent"
	"github.com/EshkinKot1980/metrics/internal/agent/client"
	oldClient "github.com/EshkinKot1980/metrics/internal/agent/client/compatible"
	"github.com/EshkinKot1980/metrics/internal/agent/monitor"
	"github.com/EshkinKot1980/metrics/internal/agent/storage"
)

type reporter interface {
	Report()
}

func main() {
	cfg := agent.MustLoadConfig()
	s := storage.New()
	m := monitor.New(s)
	am := monitor.NewAdditionalMonitor(s)

	var r reporter
	if cfg.BatchReport {
		r = client.New(s, cfg.BaseURL, cfg.SecretKey)
	} else {
		r = oldClient.New(s, cfg.BaseURL, cfg.RateLimit)
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
