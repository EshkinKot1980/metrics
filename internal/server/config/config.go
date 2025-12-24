// Модуль конфигурации сервера.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

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

// Временный конфиг, содержащий интервал из конфигурационного файла
type TmpConfig struct {
	Interval CustomDuration `json:"store_interval"`
}

// Конгигурация файлового хранилища.
type FileStorageConfig struct {
	// Интервал сохранения данных в секундах
	Interval time.Duration `env:"STORE_INTERVAL" env-default:"0s"`
	// Путь файла сохранения данных, по умолчанию "data/server/metrics.json"
	Path string `json:"store_file" env:"STORE_FILE" env-default:"data/server/metrics.json"`
	// Определяет нужно ли загружать данные из файла при запуске приложения.
	Restore bool `json:"restore" env:"RESTORE" env-default:"false"`
}

// Конфигурация сервера.
type Config struct {
	// DSN для подключения к СУБД Postgres.
	DatabaseDSN string `json:"database_dsn" env:"DATABASE_DSN"`
	// Адрес для работы веб вервера в формате "host:port".
	ServerAddr string `json:"address" env:"ADDRESS" env-default:"localhost:8080"`
	// Ключ для проверки подписи запросов и подписи ответов,
	// если не указан то проверка и подпись не производятся.
	SecretKey string `env:"KEY"`
	// Путь к приватному ключу для расшифровки запросов
	// Если не указан расшифрока не производится
	PrivateKey string `json:"crypto_key" env:"CRYPTO_KEY"`
	// Файл для сохранения событий аудита.
	AuditFile string `env:"AUDIT_FILE"`
	// URL для отправки событий аудита.
	AuditURL string `env:"AUDIT_URL"`
	// Доверенная сеть в формате CIDR, из которой можно принимать метрики,
	// если не указана, то можно принимать из любой сети
	TrustedSubnet string `json:"trusted_subnet" env:"TRUSTED_SUBNET" env-default:""`
	FileCfg       FileStorageConfig
}

// Загружает конфигурацию из флагов и переменных среды. Приоритет имеют переменные среды.
func Load() (*Config, error) {
	var (
		cfg    = &Config{}
		tmpCfg = &TmpConfig{}
		err    error
	)

	flag.CommandLine = flag.NewFlagSet("", flag.ContinueOnError)

	var (
		flagC         = flag.String("c", "", "config file path")
		flagConfig    = flag.String("config", "", "config file path")
		flagA         = flag.String("a", "localhost:8080", "address to serve")
		flagD         = flag.String("d", "", "database dsn")
		flagF         = flag.String("f", "data/server/metrics.json", "file storage path")
		flagK         = flag.String("k", "", "secret key")
		flagI         = flag.Uint64("i", 300, "store interval in seconds")
		flagR         = flag.Bool("r", false, "restore server state from file on start")
		flagT         = flag.String("t", "", "trusted subnet in cidr format")
		flagAuditFile = flag.String("audit-file", "", "audit file path")
		flagAuditURL  = flag.String("audit-url", "", "audit url")
		flagCryptoKey = flag.String("crypto-key", "", "private key path")
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
		cfg.FileCfg.Interval = time.Duration(tmpCfg.Interval)

		err = cleanenv.ReadConfig(configPath, &cfg.FileCfg)
		if err != nil {
			return nil, fmt.Errorf("failed to parse config file %w", err)
		}
	}

	// Применяем заданные флаги (средний приоритет)
	flag.Visit(func(fl *flag.Flag) {
		switch fl.Name {
		case "a":
			cfg.ServerAddr = *flagA
		case "d":
			cfg.DatabaseDSN = *flagD
		case "k":
			cfg.SecretKey = *flagK
		case "crypto-key":
			cfg.PrivateKey = *flagCryptoKey
		case "audit-file":
			cfg.AuditFile = *flagAuditFile
		case "audit-url":
			cfg.AuditURL = *flagAuditURL
		case "i":
			cfg.FileCfg.Interval = time.Duration(*flagI) * time.Second
		case "f":
			cfg.FileCfg.Path = *flagF
		case "r":
			cfg.FileCfg.Restore = *flagR
		case "t":
			cfg.TrustedSubnet = *flagT
		}
	})

	// Применяем переменные окружения (высший приоритет)
	err = cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to read environment variables %w", err)
	}
	err = cleanenv.ReadEnv(&cfg.FileCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to read environment variables %w", err)
	}

	return cfg, nil
}
