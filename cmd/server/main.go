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
	r.Get("/update/{type}/{name}/{value}", metricServer.UpdateHandler)
	//http.HandleFunc("/update/", metricServer.UpdateHandler)

	// Запуск сервера
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
