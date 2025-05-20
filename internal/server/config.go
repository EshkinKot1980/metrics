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
	ServerAddr string
	StorageCfg FileStorageConfig
}

func MustLoadConfig() *Config {
	var (
		a, f string
		i    uint64
		r    bool
		err  error
	)

	flag.StringVar(&a, "a", "localhost:8080", "address to serve")
	flag.StringVar(&f, "f", "storage/server/metrics.json", "file storage path")
	flag.Uint64Var(&i, "i", 300, "store interval in seconds")
	flag.BoolVar(&r, "r", false, "restore server state from file on start")

	flag.Parse()

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		a = envAddr
	}

	if envPath := os.Getenv("FILE_STORAGE_PATH"); envPath != "" {
		f = envPath
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
		ServerAddr: a,
		StorageCfg: FileStorageConfig{
			Interval: i,
			Path:     f,
			Restore:  r,
		},
	}
}
