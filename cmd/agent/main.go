package main

import (
	"crypto/rsa"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/EshkinKot1980/metrics/internal/agent"
	"github.com/EshkinKot1980/metrics/internal/agent/client"
	oldAPIclient "github.com/EshkinKot1980/metrics/internal/agent/client/compatible"
	"github.com/EshkinKot1980/metrics/internal/agent/monitor"
	"github.com/EshkinKot1980/metrics/internal/agent/storage"
	"github.com/EshkinKot1980/metrics/internal/common/utils"
)

type reporter interface {
	Report()
}

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

	var publicKey *rsa.PublicKey
	if cfg.PublicKey != "" {
		publicKey, err = utils.LoadPublicKey(cfg.PublicKey)
		if err != nil {
			log.Fatal("failed to load public key: ", err)
		}
	}

	var r reporter
	if cfg.BatchReport {
		r = client.New(s, cfg.BaseURL, cfg.SecretKey, publicKey)
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
