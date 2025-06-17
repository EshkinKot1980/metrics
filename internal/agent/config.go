package agent

import (
	"errors"
	"flag"
	"log"
	"os"
	"strconv"
)

var ErrNotNaturalNumber = errors.New("the value must be a natural number")

type Config struct {
	BaseURL        string
	PollInterval   uint64
	ReportInterval uint64
	RateLimit      uint64
	SecretKey      string
}

func MustLoadConfig() *Config {
	var (
		schema    = "http"
		addr, key string
	)

	pollInterval := new(natural)
	pollInterval.value = 2
	reportInterval := new(natural)
	reportInterval.value = 10
	rateLimit := new(natural)
	rateLimit.value = 10

	flag.StringVar(&addr, "a", "localhost:8080", "server address")
	flag.StringVar(&key, "k", "", "secret key")
	flag.Var(pollInterval, "p", "poll interval in seconds")
	flag.Var(reportInterval, "r", "report interval in seconds")
	flag.Var(rateLimit, "l", "rate limit, limit of simultaneous requests")

	flag.Parse()

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		addr = envAddr
	}

	if envKey := os.Getenv("KEY"); envKey != "" {
		key = envKey
	}

	if envPI := os.Getenv("POLL_INTERVAL"); envPI != "" {
		err := pollInterval.Set(envPI)
		if err != nil {
			log.Fatal(err)
		}
	}

	if envRI := os.Getenv("REPORT_INTERVALL"); envRI != "" {
		err := reportInterval.Set(envRI)
		if err != nil {
			log.Fatal(err)
		}
	}

	if envRL := os.Getenv("RATE_LIMIT"); envRL != "" {
		err := rateLimit.Set(envRL)
		if err != nil {
			log.Fatal(err)
		}
	}

	return &Config{
		BaseURL:        schema + "://" + addr,
		PollInterval:   pollInterval.value,
		ReportInterval: reportInterval.value,
		RateLimit:      rateLimit.value,
		SecretKey:      key,
	}
}

type natural struct {
	value uint64
}

func (n *natural) String() string {
	return strconv.FormatUint(n.value, 10)
}

func (n *natural) Set(flagValue string) error {
	v, err := strconv.ParseUint(flagValue, 10, 64)
	if err != nil {
		return ErrNotNaturalNumber
	}
	if v == 0 {
		return ErrNotNaturalNumber
	}
	n.value = v
	return nil
}
