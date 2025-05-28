package config

import (
	"flag"
	"github.com/caarlos0/env/v11"
	"strings"
)

const (
	DefaultServerAddr     = "localhost:8080"
	DefaultReportInterval = int64(10)
	DefaultPollInterval   = int64(2)
	DefaultUseGzip        = bool(true)
)

type Config struct {
	ServerAddr     string `env:"ADDRESS"`
	PollInterval   int64  `env:"POLL_INTERVAL"`
	ReportInterval int64  `env:"REPORT_INTERVAL"`
	UseGzip        bool   `env:"USE_GZIP" envDefault:"true"`
}

// выставляет значения конфигу из аргументов командной строки
func SetConfig() (Config, error) {
	var cfg Config

	err := env.Parse(&cfg)
	if err != nil {
		return cfg, err
	}

	a := flag.String("a", DefaultServerAddr, "Адрес сервера")
	p := flag.Int64("p", DefaultPollInterval, "Интервал сбора метрик")
	r := flag.Int64("r", DefaultReportInterval, "Интервал отправки метрик")
	flag.Parse()

	if strings.TrimSpace(cfg.ServerAddr) == "" {
		cfg.ServerAddr = *a
	}

	if cfg.PollInterval == 0 {
		cfg.PollInterval = *p
	}

	if cfg.ReportInterval == 0 {
		cfg.ReportInterval = *r
	}

	return cfg, err
}
