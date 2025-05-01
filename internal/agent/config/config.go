package config

import (
	"flag"
	"os"
	"strconv"
)

// TODO: добавть настройки http-клиента
type Config struct {
	BaseURL        string
	PollInterval   uint64
	ReportInterval uint64
}

func MustLoad() *Config {
	var (
		schema = "http"
		addr   string
		pi, ri uint64
		err    error
	)

	flag.StringVar(&addr, "a", "localhost:8080", "address to serve")
	flag.Uint64Var(&pi, "p", 2, "poll interval in seconds")
	flag.Uint64Var(&ri, "r", 10, "report interval in seconds")

	flag.Parse()

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		addr = envAddr
	}

	if envPI := os.Getenv("POLL_INTERVAL"); envPI != "" {
		pi, err = strconv.ParseUint(envPI, 10, 64)
		if err != nil {
			panic(err)
		}
	}

	if envRI := os.Getenv("REPORT_INTERVALL"); envRI != "" {
		pi, err = strconv.ParseUint(envRI, 10, 64)
		if err != nil {
			panic(err)
		}
	}

	return &Config{
		BaseURL:        schema + "://" + addr,
		PollInterval:   pi,
		ReportInterval: ri,
	}
}
