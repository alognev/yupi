package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
	"yupi/internal/config"
	"yupi/internal/httptransport/handlers"
	"yupi/internal/httptransport/middlewares"
	"yupi/internal/logger"
	"yupi/internal/repository"
	"yupi/internal/service/server"
)

func main() {
	if err := logger.Initialize("info"); err != nil {
		log.Fatal("Не удалось инициировать логгер")
	}
	// Инициализация конфига
	cfg := config.SetServerConfig()
	// Инициализация хранилища
	storage := repository.NewMemStorage()
	fileStorage := repository.NewFileStorage(storage)

	// Инициализация сервера метрик, отдельно разбит на хендлер с хранилищем метрик и отдельно на сохранялку в файл
	metricHandler := handlers.NewMetricServer(storage)
	metricFileServer := server.NewMetricsSaver(*metricHandler, *fileStorage, &cfg)
	err := metricFileServer.Run()

	if err != nil {
		log.Fatal("Не удалось запустить обработчик файлов")
	}

	// Инициализация роутера
	r := chi.NewRouter()
	r.Use(logger.LoggingRequestMiddleware, middlewares.GzipMiddleware)

	// Настройка маршрутов

	r.With(middleware.AllowContentType("application/json")).
		Post("/update/", metricHandler.JSONUpdateHandler)

	r.With(middleware.AllowContentType("application/json")).
		Post("/value/", metricHandler.JSONValueHandler)

	r.Post("/update/{type}/{name}/{value}", metricHandler.UpdateHandler)
	r.Get("/value/{type}/{name}", metricHandler.ValueHandler)
	r.Get("/", metricHandler.MainHandler)

	// Запуск сервера
	log.Println("Starting server on " + cfg.ServerAddr)
	log.Fatal(http.ListenAndServe(cfg.ServerAddr, r))
}
