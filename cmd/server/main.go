package main

import (
	"flag"
	"github.com/caarlos0/env/v11"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
	"strings"
	"yupi/internal/config"
	"yupi/internal/httptransport/handlers"
	"yupi/internal/httptransport/middlewares"
	"yupi/internal/logger"
	"yupi/internal/repository"
)

type Config struct {
	ServerAddr string `env:"ADDRESS"`
}

func main() {
	if err := logger.Initialize("info"); err != nil {
		log.Fatal("Не удалось инициировать логгер")
	}
	// Инициализация конфига
	cfg := setConfig()
	// Инициализация хранилища
	storage := repository.NewMemStorage()

	// Инициализация сервера метрик
	metricServer := handlers.NewMetricServer(storage)

	// Инициализация роутера
	r := chi.NewRouter()
	r.Use(logger.LoggingRequestMiddleware, middlewares.GzipMiddleware)

	// Настройка маршрутов

	r.With(middleware.AllowContentType("application/json")).
		Post("/update/", metricServer.JSONUpdateHandler)

	r.With(middleware.AllowContentType("application/json")).
		Post("/value/", metricServer.JSONValueHandler)

	r.Post("/update/{type}/{name}/{value}", metricServer.UpdateHandler)
	r.Get("/value/{type}/{name}", metricServer.ValueHandler)
	r.Get("/", metricServer.MainHandler)

	// Запуск сервера
	log.Println("Starting server on " + cfg.ServerAddr)
	log.Fatal(http.ListenAndServe(cfg.ServerAddr, r))
}

// Выставляет значения конфиг из аргументов командной строки
func setConfig() Config {
	var cfg Config

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	a := flag.String("a", config.DefaultServerAddr, "Адрес сервера")
	flag.Parse()

	if strings.TrimSpace(cfg.ServerAddr) == "" {
		cfg.ServerAddr = *a
	}

	return cfg
}
