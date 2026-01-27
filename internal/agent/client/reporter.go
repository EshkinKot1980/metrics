package client

import (
	"crypto/rsa"
	"fmt"
	"net"

	"github.com/EshkinKot1980/metrics/internal/agent"
	"github.com/EshkinKot1980/metrics/internal/agent/client/grpc"
	"github.com/EshkinKot1980/metrics/internal/agent/client/http"
	"github.com/EshkinKot1980/metrics/internal/common/utils"
)

type Reporter interface {
	Report()
}

func NewReporter(cfg *agent.Config, s http.Storage) (Reporter, error) {
	if cfg.GRPCaddr != "" {
		ip, err := defineIP()
		if err != nil {
			return nil, fmt.Errorf("failed to define agent ip address: %w", err)
		}

		return grpc.NewClient(s, cfg.GRPCaddr, ip), nil
	}

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

		ip, err := defineIP()
		if err != nil {
			return nil, fmt.Errorf("failed to define agent ip address: %w", err)
		}

		return http.NewBatchClient(s, baseURL, cfg.SecretKey, publicKey, ip), nil
	}

	return http.NewMultithreadedClient(s, baseURL, cfg.RateLimit), nil
}

func defineIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err

	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}

	return "", fmt.Errorf("ip adress not found")
}
