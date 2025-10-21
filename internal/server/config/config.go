// Модуль конфигурации сервера.
package config

import (
	"flag"
	"log"
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
	// Файл для сохранения событий аудита.
	AuditFile string
	// URL для отправки событий аудита.
	AuditURL string
	FileCfg  FileStorageConfig
}

// Загружает конфигурацию из флагов и переменных среды. Приоритет имеют переменные среды.
// Может вызывать log.Fatal(), поэтому вызывается только в начале инициализации приложения.
func MustLoad() *Config {
	var (
		a, d, f, k string
		i          uint64
		r          bool
		auditFile  string
		auditURL   string
		err        error
	)

	flag.StringVar(&a, "a", "localhost:8080", "address to serve")
	flag.StringVar(&d, "d", "", "database dsn")
	flag.StringVar(&f, "f", "data/server/metrics.json", "file storage path")
	flag.StringVar(&k, "k", "", "secret key")
	flag.Uint64Var(&i, "i", 300, "store interval in seconds")
	flag.BoolVar(&r, "r", false, "restore server state from file on start")
	flag.StringVar(&auditFile, "audit-file", "", "audit file path")
	flag.StringVar(&auditURL, "audit-url", "", "audit url")

	flag.Parse()

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

	if envInterval := os.Getenv("STORE_INTERVAL"); envInterval != "" {
		i, err = strconv.ParseUint(envInterval, 10, 64)
		if err != nil {
			log.Fatal(err)
		}
	}

	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		r, err = strconv.ParseBool(envRestore)
		if err != nil {
			log.Fatal(err)
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

	return &Config{
		DatabaseDSN: d,
		ServerAddr:  a,
		SecretKey:   k,
		AuditFile:   auditFile,
		AuditURL:    auditURL,
		FileCfg: FileStorageConfig{
			Interval: i,
			Path:     f,
			Restore:  r,
		},
	}
}
