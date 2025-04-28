package main

import (
	"flag"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"yupi/internal/httptransport/handlers"
	"yupi/internal/repository"
)

type Config struct {
	ServerAddr string
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
	var config Config

	flag.StringVar(&config.ServerAddr, "a", "http://localhost:8080", "Адрес сервера")
	flag.Parse()

	return config
}
