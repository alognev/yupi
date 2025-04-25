package main

import (
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

	// Настройка маршрутов
	http.HandleFunc("/update/", metricServer.UpdateHandler)

	// Запуск сервера
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
