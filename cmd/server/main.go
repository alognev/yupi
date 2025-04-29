package main

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"strings"
	"yupi/internal/config"
	"yupi/internal/httptransport/handlers"
	"yupi/internal/repository"
)

type Config struct {
	ServerAddr string `env:"ADDRESS"`
}

func main() {
	// Инициализация конфига
	config := setConfig()
	// Инициализация хранилища
	storage := repository.NewMemStorage()

	// Инициализация сервера метрик
	metricServer := handlers.NewMetricServer(storage)

	// Инициализация роутера
	r := chi.NewRouter()

	// Настройка маршрутов
	r.Post("/update/{type}/{name}/{value}", metricServer.UpdateHandler)
	r.Get("/value/{type}/{name}", metricServer.ValueHandler)
	r.Get("/", metricServer.MainHandler)

	// Запуск сервера
	log.Println("Starting server on " + config.ServerAddr)
	log.Fatal(http.ListenAndServe(config.ServerAddr, r))
}

// выставляет значения конфигу из аргументов командной строки
func setConfig() Config {
	var cfg Config

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	if strings.TrimSpace(cfg.ServerAddr) == "" {
		flag.StringVar(&cfg.ServerAddr, "a", config.DefaultServerAddr, "Адрес сервера")
	}

	flag.Parse()
	return cfg
}
