package server

import (
	"flag"
	"log"
	"os"
	"strconv"
)

type FileStorageConfig struct {
	Interval uint64
	Path     string
	Restore  bool
}

// TODO: добавть настройки http-сервера
type Config struct {
	DatabaseDSN string
	ServerAddr  string
	SecretKey   string
	FileCfg     FileStorageConfig
}

func MustLoadConfig() *Config {
	var (
		a, d, f, k string
		i          uint64
		r          bool
		err        error
	)

	flag.StringVar(&a, "a", "localhost:8080", "address to serve")
	flag.StringVar(&d, "d", "", "database dsn")
	flag.StringVar(&f, "f", "data/server/metrics.json", "file storage path")
	flag.StringVar(&k, "k", "", "secret key")
	flag.Uint64Var(&i, "i", 300, "store interval in seconds")
	flag.BoolVar(&r, "r", false, "restore server state from file on start")

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

	return &Config{
		DatabaseDSN: d,
		ServerAddr:  a,
		SecretKey:   k,
		FileCfg: FileStorageConfig{
			Interval: i,
			Path:     f,
			Restore:  r,
		},
	}
}
