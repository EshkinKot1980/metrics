// Модуль agent реализует клиентскую часть приложения сбора метрик.
package agent

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
)

var ErrNotNaturalNumber = errors.New("the value must be a natural number")

// Конфигурация агента.
type Config struct {
	BaseURL        string // корневой путь API сервера
	BatchReport    bool   // указывает ну жли ли отправлять все метрики одним запросом
	PollInterval   uint64 // интервал сбора мертик в секундах
	ReportInterval uint64 // интервал отправки данных на сервер в секундах
	RateLimit      uint64 // количество одновременных запросов к серверу
	SecretKey      string // ключ для подписи запросов
	PublicKey      string // ключ для шифрования запросов
	PprofAdrr      string // адрес профилировщика в формате "host:port"
}

// Загружает конфигурацию из флагов и переменных среды. Приоритет имеют переменные среды.
func LoadConfig() (*Config, error) {
	var (
		schema      = "http"
		addr, key   string
		pprofAdrr   string
		cryptoKey   string
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

	flag.CommandLine = flag.NewFlagSet("", flag.ContinueOnError)

	flag.StringVar(&addr, "a", "localhost:8080", "server address")
	flag.StringVar(&key, "k", "", "secret key")
	flag.StringVar(&cryptoKey, "crypto-key", "", "public key path")
	flag.Var(pollInterval, "p", "poll interval in seconds")
	flag.Var(reportInterval, "r", "report interval in seconds")
	flag.Var(rateLimit, "l", "rate limit, limit of simultaneous requests")
	flag.BoolVar(&batchReport, "b", true, "batch report, send all metrics in one request")
	flag.StringVar(&pprofAdrr, "pprof-addr", "", pprofUsage)

	err = flag.CommandLine.Parse(os.Args[1:])
	if err != nil {
		return &Config{}, fmt.Errorf("failed to parse flags %w", err)
	}

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		addr = envAddr
	}

	if envKey := os.Getenv("KEY"); envKey != "" {
		key = envKey
	}

	if envCryptoKey := os.Getenv("CRYPTO_KEY"); envCryptoKey != "" {
		cryptoKey = envCryptoKey
	}

	if envPI := os.Getenv("POLL_INTERVAL"); envPI != "" {
		err = pollInterval.Set(envPI)
		if err != nil {
			return &Config{}, fmt.Errorf("failed to parse POLL_INTERVAL %w", err)
		}
	}

	if envRI := os.Getenv("REPORT_INTERVALL"); envRI != "" {
		err = reportInterval.Set(envRI)
		if err != nil {
			return &Config{}, fmt.Errorf("failed to parse REPORT_INTERVALL %w", err)
		}
	}

	if envRL := os.Getenv("RATE_LIMIT"); envRL != "" {
		err = rateLimit.Set(envRL)
		if err != nil {
			return &Config{}, fmt.Errorf("failed to parse RATE_LIMIT %w", err)
		}
	}

	if envBR := os.Getenv("BATCH_REPORT"); envBR != "" {
		batchReport, err = strconv.ParseBool(envBR)
		if err != nil {
			return &Config{}, fmt.Errorf("failed to parse BATCH_REPORT %w", err)
		}
	}

	if envPprofAdrr := os.Getenv("PPROF_ADDRESS"); envPprofAdrr != "" {
		pprofAdrr = envPprofAdrr
	}

	cfg := &Config{
		BaseURL:        schema + "://" + addr,
		BatchReport:    batchReport,
		PollInterval:   pollInterval.value,
		ReportInterval: reportInterval.value,
		RateLimit:      rateLimit.value,
		SecretKey:      key,
		PublicKey:      cryptoKey,
		PprofAdrr:      pprofAdrr,
	}

	return cfg, nil
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
