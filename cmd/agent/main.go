package main

import (
	"github.com/EshkinKot1980/metrics/internal/agent/monitor"
	"github.com/EshkinKot1980/metrics/internal/agent/client"
	"github.com/EshkinKot1980/metrics/internal/agent/storage"
)

func main() {
	s := storage.New()
	go func() {
		monitor.Run(2, s)
	}()

	client.Run(10, s)
}
