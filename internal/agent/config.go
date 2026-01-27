// Модуль agent реализует клиентскую часть приложения сбора метрик.
package agent

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

var ErrNotNaturalNumber = errors.New("the value must be a natural number")

// Конфигурация агента.
type Config struct {
	// Адрес http сервера в формате "host:port"
	APIAddres string `json:"address" env:"ADDRESS" env-default:"localhost:8080"`
	// Адрес grpc cервера в формате "host:port", если не указа метрики отправляются на http сервер
	GRPCaddr string `json:"grpc_address" env:"GRPC_ADDRESS" env-default:""`
	// Указывает ну жли ли отправлять все метрики одним запросом
	BatchReport bool `env:"BATCH_REPORT" env-default:"true"`
	// Интервал сбора мертик
	PollInterval time.Duration `env:"POLL_INTERVAL" env-default:"2s"`
	// Интервал отправки данных на сервер
	ReportInterval time.Duration `env:"REPORT_INTERVALL" env-default:"10s"`
	// Количество одновременных запросов к серверу RATE_LIMIT
	RateLimit uint64 `env:"RATE_LIMIT" env-default:"10"`
	// Ключ для подписи запросов, если не указан, то подпись не производится
	SecretKey string `env:"KEY" env-default:""`
	// Путь к публичному ключу для шифрования запросов
	// Если не указан,  шифрования не производится
	PublicKey string `json:"crypto_key" env:"CRYPTO_KEY" env-default:""`
	// Адрес профилировщика в формате "host:port", если не указан, профайлер не запускается
	// Если указать ":8080", профайлер будет доступен по адресу http://localhost:8080/debug/pprof/
	PprofAdrr string `env:"PPROF_ADDRESS" env-default:""`
}

// CustomDuration необходим для корректного преобразования формата "10m13s" в json файле в time.Duration
type CustomDuration time.Duration

func (cd *CustomDuration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*cd = CustomDuration(d)
	return nil
}

// Временный конфиг, содержащий интервалs из конфигурационного файла
type TmpConfig struct {
	PollInterval   CustomDuration `json:"poll_interval"`
	ReportInterval CustomDuration `json:"report_interval"`
}

// Загружает конфигурацию из флагов и переменных среды. Приоритет имеют переменные среды.
func LoadConfig() (*Config, error) {
	var (
		cfg    = &Config{}
		tmpCfg = &TmpConfig{}
		err    error
	)

	flag.CommandLine = flag.NewFlagSet("", flag.ContinueOnError)

	var (
		flagC         = flag.String("c", "", "config file path")
		flagConfig    = flag.String("config", "", "config file path")
		flagA         = flag.String("a", "localhost:8080", "http server address")
		flagB         = flag.Bool("b", true, "batch report, send all metrics in one request")
		flagG         = flag.String("g", "", "grpc server address")
		flagK         = flag.String("k", "", "secret key")
		flagL         = flag.Uint64("l", 10, "rate limit, limit of simultaneous requests")
		flagP         = flag.Uint64("p", 2, "poll interval in seconds")
		flagR         = flag.Uint64("r", 10, "report interval in seconds")
		flagCryptoKey = flag.String("crypto-key", "", "public key path")
		flagPprofAdrr = flag.String("pprof-addr", "", "profiler address:port")
	)

	err = flag.CommandLine.Parse(os.Args[1:])
	if err != nil {
		return nil, fmt.Errorf("failed to parse flags %w", err)
	}

	// Разбираем файл конфигурации (низший приоритет)
	configPath := *flagC
	if configPath == "" {
		configPath = *flagConfig
	}
	if configPath != "" {
		err = cleanenv.ReadConfig(configPath, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to parse config file %w", err)
		}

		// танцы с бубнов вокруг парcинга time.duration из json
		err = cleanenv.ReadConfig(configPath, tmpCfg)
		if err != nil {
			return nil, fmt.Errorf("failed to parse config file %w", err)
		}
		cfg.PollInterval = time.Duration(tmpCfg.PollInterval)
		cfg.ReportInterval = time.Duration(tmpCfg.ReportInterval)
	}

	// Применяем заданные флаги (средний приоритет)
	flag.Visit(func(fl *flag.Flag) {
		switch fl.Name {
		case "a":
			cfg.APIAddres = *flagA
		case "d":
			cfg.BatchReport = *flagB
		case "g":
			cfg.GRPCaddr = *flagG
		case "k":
			cfg.SecretKey = *flagK
		case "l":
			cfg.RateLimit = *flagL
		case "p":
			cfg.PollInterval = time.Duration(*flagP) * time.Second
		case "r":
			cfg.ReportInterval = time.Duration(*flagR) * time.Second
		case "crypto-key":
			cfg.PublicKey = *flagCryptoKey
		case "pprof-addr":
			cfg.PprofAdrr = *flagPprofAdrr

		}
	})

	// Применяем переменные окружения (высший приоритет)
	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to read environment variables %w", err)
	}

	// Проверям интервалы и RateLimit
	if cfg.PollInterval <= 0 {
		return nil, fmt.Errorf("invalid poll interval: %w", ErrNotNaturalNumber)
	}
	if cfg.ReportInterval <= 0 {
		return nil, fmt.Errorf("invalid report interval: %w", ErrNotNaturalNumber)
	}
	if cfg.RateLimit == 0 {
		return nil, fmt.Errorf("invalid rate limit: %w", ErrNotNaturalNumber)
	}

	return cfg, nil
}
