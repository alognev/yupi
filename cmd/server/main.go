package main

import (
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"yupi/internal/httptransport/handlers"
	"yupi/internal/repository"
)

func main() {
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
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
