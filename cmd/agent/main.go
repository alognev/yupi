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
		config.PollInterval,
		config.ReportInterval,
	)
	agent.Run()
}

// выставляет значения конфигу из аргументов командной строки
func setConfig() Config {
	var config Config

	flag.StringVar(&config.ServerAddr, "a", "localhost:8080", "Адрес сервера")
	p := flag.Int64("p", 2, "Адрес сервера")
	r := flag.Int64("r", 10, "Адрес сервера")
	//flag.DurationVar(&config.PollInterval, "pp", 2, "Интервал сбора метрик")
	//flag.DurationVar(&config.ReportInterval, "rr", 10, "Интервал отправки метрик")

	config.PollInterval = time.Duration(*p) * time.Second
	config.ReportInterval = time.Duration(*r) * time.Second

	flag.Parse()

	return config
}
