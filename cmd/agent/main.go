package main

import (
	"flag"
	"time"
	"yupi/internal/service/agent"
)

type Config struct {
	ServerAddr     string
	PollInterval   time.Duration
	ReportInterval time.Duration
}

func main() {
	config := setConfig()
	agent := agent.NewAgent(
		config.ServerAddr,
		config.PollInterval*time.Second,
		config.ReportInterval*time.Second,
	)
	agent.Run()
}

// выставляет значения конфигу из аргументов командной строки
func setConfig() Config {
	var config Config

	flag.StringVar(&config.ServerAddr, "a", "http://localhost:8080", "Адрес сервера")
	flag.DurationVar(&config.PollInterval, "p", 2, "Интервал сбора метрик")
	flag.DurationVar(&config.ReportInterval, "r", 10, "Интервал отправки метрик")
	flag.Parse()

	return config
}
