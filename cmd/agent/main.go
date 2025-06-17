package main

import (
	"time"

	"github.com/EshkinKot1980/metrics/internal/agent"
	"github.com/EshkinKot1980/metrics/internal/agent/client"
	"github.com/EshkinKot1980/metrics/internal/agent/monitor"
	"github.com/EshkinKot1980/metrics/internal/agent/storage"
)

func main() {
	cfg := agent.MustLoadConfig()
	s := storage.New()
	c := client.New(s, cfg.BaseURL, cfg.SecretKey)
	m := monitor.New(s)
	am := monitor.NewAdditionalMonitor(s)
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
		c.Report()
	}
}
