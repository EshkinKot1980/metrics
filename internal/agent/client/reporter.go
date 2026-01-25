package client

import (
	"crypto/rsa"
	"fmt"

	"github.com/EshkinKot1980/metrics/internal/agent"
	"github.com/EshkinKot1980/metrics/internal/agent/client/http"
	"github.com/EshkinKot1980/metrics/internal/common/utils"
)

type Reporter interface {
	Report()
}

func NewReporter(cfg *agent.Config, s http.Storage) (Reporter, error) {
	baseURL := "http://" + cfg.APIAddres
	if cfg.BatchReport {
		var publicKey *rsa.PublicKey
		var err error

		if cfg.PublicKey != "" {
			publicKey, err = utils.LoadPublicKey(cfg.PublicKey)
			if err != nil {
				return nil, fmt.Errorf("failed to load public key: %w", err)
			}
		}

		return http.NewBatchClient(s, baseURL, cfg.SecretKey, publicKey)
	}

	return http.NewMultithreadedClient(s, baseURL, cfg.RateLimit), nil
}
