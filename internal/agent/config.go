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
	BatchReport    bool
	PollInterval   uint64
	ReportInterval uint64
	RateLimit      uint64
	SecretKey      string
	PprofAdrr      string
}

func MustLoadConfig() *Config {
	var (
		schema      = "http"
		addr, key   string
		pprofAdrr   string
		batchReport bool
		err         error
	)

	pollInterval := new(natural)
	pollInterval.value = 2
	reportInterval := new(natural)
	reportInterval.value = 10
	rateLimit := new(natural)
	rateLimit.value = 10

	pprofUsage := "profiler address:port, if specify :8080," +
		" profiler will be available at http://localhost:8080/debug/pprof/"

	flag.StringVar(&addr, "a", "localhost:8080", "server address")
	flag.StringVar(&key, "k", "", "secret key")
	flag.Var(pollInterval, "p", "poll interval in seconds")
	flag.Var(reportInterval, "r", "report interval in seconds")
	flag.Var(rateLimit, "l", "rate limit, limit of simultaneous requests")
	flag.BoolVar(&batchReport, "b", true, "batch report, send all metrics in one request")
	flag.StringVar(&pprofAdrr, "pprof-addr", "", pprofUsage)

	flag.Parse()

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		addr = envAddr
	}

	if envKey := os.Getenv("KEY"); envKey != "" {
		key = envKey
	}

	if envPI := os.Getenv("POLL_INTERVAL"); envPI != "" {
		err = pollInterval.Set(envPI)
		if err != nil {
			log.Fatal(err)
		}
	}

	if envRI := os.Getenv("REPORT_INTERVALL"); envRI != "" {
		err = reportInterval.Set(envRI)
		if err != nil {
			log.Fatal(err)
		}
	}

	if envRL := os.Getenv("RATE_LIMIT"); envRL != "" {
		err = rateLimit.Set(envRL)
		if err != nil {
			log.Fatal(err)
		}
	}

	if envBR := os.Getenv("BATCH_REPORT"); envBR != "" {
		batchReport, err = strconv.ParseBool(envBR)
		if err != nil {
			log.Fatal(err)
		}
	}

	if envPprofAdrr := os.Getenv("PPROF_ADDRESS"); envPprofAdrr != "" {
		pprofAdrr = envPprofAdrr
	}

	return &Config{
		BaseURL:        schema + "://" + addr,
		BatchReport:    batchReport,
		PollInterval:   pollInterval.value,
		ReportInterval: reportInterval.value,
		RateLimit:      rateLimit.value,
		SecretKey:      key,
		PprofAdrr:      pprofAdrr,
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
