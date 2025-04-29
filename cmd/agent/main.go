package main

import (
	"flag"
	"github.com/caarlos0/env/v11"
	"log"
	"strings"
	"yupi/internal/config"
	"yupi/internal/service/agent"
)

type Config struct {
	ServerAddr     string `env:"ADDRESS"`
	PollInterval   int64  `env:"POLL_INTERVAL"`
	ReportInterval int64  `env:"REPORT_INTERVAL"`
}

func main() {
	cfg := setConfig()
	myAgent := agent.NewAgent(
		cfg.ServerAddr,
		cfg.PollInterval,
		cfg.ReportInterval,
	)
	myAgent.Run()
}

// выставляет значения конфигу из аргументов командной строки
func setConfig() Config {
	var cfg Config

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	a := flag.String("a", config.DefaultServerAddr, "Адрес сервера")
	p := flag.Int64("p", config.DefaultPollInterval, "Интервал сбора метрик")
	r := flag.Int64("r", config.DefaultReportInterval, "Интервал отправки метрик")
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

	return cfg
}
