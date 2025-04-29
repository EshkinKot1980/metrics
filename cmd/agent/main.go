package main

import (
	"flag"
	"time"
	"github.com/EshkinKot1980/metrics/internal/agent/monitor"
	"github.com/EshkinKot1980/metrics/internal/agent/client"
	"github.com/EshkinKot1980/metrics/internal/agent/storage"
)

type Config struct {
	schema         string
	address 	   string
	URL			   string
	PollInterval   int
	ReportInterval int
}

func main() {
	cfg := loadConfig()
	s := storage.New()
	c := client.New(s, cfg.URL)
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

func loadConfig() *Config {
	schema := "http"
	addr := flag.String("a", "localhost:8080", "address to serve in format host:port")
	pi := flag.Int("p", 2, "poll interval in seconds")
	ri := flag.Int("r", 10, "report interval in seconds")
	

	flag.Parse()

	return &Config{
		schema:  "http",
		address: *addr,
		URL: schema + "://" + *addr,
		PollInterval: *pi,
		ReportInterval: *ri, 
	}
}