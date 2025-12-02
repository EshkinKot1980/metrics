// Модуль конфигурации сервера.
package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

// Конгигурация файлового хранилища.
type FileStorageConfig struct {
	// Интервал сохранения данных в секундах
	Interval uint64
	// Путь файла сохранения данных, по умолчанию "data/server/metrics.json"
	Path string
	// Определяет нужно ли загружать данные из файла при запуске приложения.
	Restore bool
}

// Конфигурация сервера.
type Config struct {
	// DSN для подключения к СУБД Postgres.
	DatabaseDSN string
	// Адрес для работы веб вервера в формате "host:port".
	ServerAddr string
	// Ключ для проверки подписи запросов и подписи ответов,
	// если не указан то проверка и подпись не производятся.
	SecretKey string
	// Путь к приватному ключу для расшифровки запросов
	// Если не указан расшифрока не производится
	PrivateKey string
	// Файл для сохранения событий аудита.
	AuditFile string
	// URL для отправки событий аудита.
	AuditURL string
	FileCfg  FileStorageConfig
}

// Загружает конфигурацию из флагов и переменных среды. Приоритет имеют переменные среды.
func Load() (*Config, error) {
	var (
		a, d, f, k string
		i          uint64
		r          bool
		auditFile  string
		auditURL   string
		cryptoKey  string
		err        error
	)

	flag.CommandLine = flag.NewFlagSet("", flag.ContinueOnError)

	flag.StringVar(&a, "a", "localhost:8080", "address to serve")
	flag.StringVar(&d, "d", "", "database dsn")
	flag.StringVar(&f, "f", "data/server/metrics.json", "file storage path")
	flag.StringVar(&k, "k", "", "secret key")
	flag.StringVar(&cryptoKey, "crypto-key", "", "private key path")
	flag.Uint64Var(&i, "i", 300, "store interval in seconds")
	flag.BoolVar(&r, "r", false, "restore server state from file on start")
	flag.StringVar(&auditFile, "audit-file", "", "audit file path")
	flag.StringVar(&auditURL, "audit-url", "", "audit url")

	err = flag.CommandLine.Parse(os.Args[1:])
	if err != nil {
		return &Config{}, fmt.Errorf("failed to parse flags %w", err)
	}

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		a = envAddr
	}

	if envDSN := os.Getenv("DATABASE_DSN"); envDSN != "" {
		d = envDSN
	}

	if envPath := os.Getenv("FILE_STORAGE_PATH"); envPath != "" {
		f = envPath
	}

	if envKey := os.Getenv("KEY"); envKey != "" {
		k = envKey
	}

	if envCryptoKey := os.Getenv("CRYPTO_KEY"); envCryptoKey != "" {
		cryptoKey = envCryptoKey
	}

	if envInterval := os.Getenv("STORE_INTERVAL"); envInterval != "" {
		i, err = strconv.ParseUint(envInterval, 10, 64)
		if err != nil {
			return &Config{}, fmt.Errorf("failed to parse STORE_INTERVAL %w", err)
		}
	}

	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		r, err = strconv.ParseBool(envRestore)
		if err != nil {
			return &Config{}, fmt.Errorf("failed to parse RESTOREL %w", err)
		}
	}

	if envKey := os.Getenv("KEY"); envKey != "" {
		k = envKey
	}

	if envAuditFile := os.Getenv("AUDIT_FILE"); envAuditFile != "" {
		auditFile = envAuditFile
	}

	if envAuditURL := os.Getenv("AUDIT_FILE"); envAuditURL != "" {
		auditURL = envAuditURL
	}

	cfg := &Config{
		DatabaseDSN: d,
		ServerAddr:  a,
		SecretKey:   k,
		PrivateKey:  cryptoKey,
		AuditFile:   auditFile,
		AuditURL:    auditURL,
		FileCfg: FileStorageConfig{
			Interval: i,
			Path:     f,
			Restore:  r,
		},
	}

	return cfg, nil
}
