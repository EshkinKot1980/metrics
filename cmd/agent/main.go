package main

import (
	"time"
	"github.com/EshkinKot1980/metrics/internal/agent/monitor"
	"github.com/EshkinKot1980/metrics/internal/agent/client"
	"github.com/EshkinKot1980/metrics/internal/agent/storage"
)

func main() {
	s := storage.New()
	c := client.New(s, "http://localhost:8080")
	m := monitor.New(s)

	go func() {
		interval := time.Duration(2) * time.Second
		for {
			<-time.After(interval)
			m.Poll()
		}
	}()

	interval := time.Duration(10) * time.Second
	for {
		<-time.After(interval)
		c.Report()
	}
}
