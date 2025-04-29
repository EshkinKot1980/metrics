package main

import (
	"time"
	"github.com/EshkinKot1980/metrics/internal/agent/config"
	"github.com/EshkinKot1980/metrics/internal/agent/monitor"
	"github.com/EshkinKot1980/metrics/internal/agent/client"
	"github.com/EshkinKot1980/metrics/internal/agent/storage"
)

func main() {
	cfg := config.MustLoad()
	s := storage.New()
	c := client.New(s, cfg.BaseURL)
	m := monitor.New(s)

	go func() {
		interval := time.Duration(cfg.PollInterval) * time.Second
		for {
			<-time.After(interval)
			m.Poll()
		}
	}()

	interval := time.Duration(cfg.ReportInterval) * time.Second
	for {
		<-time.After(interval)
		c.Report()
	}
}
