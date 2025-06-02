package server

import (
	"time"
	"yupi/internal/config"
	"yupi/internal/httptransport/middlewares"
	"yupi/internal/repository"
)

type MetricsSaver struct {
	storage  repository.FileStorage
	config   *config.ServerConfig
	stopChan chan struct{}
}

func NewMetricsSaver(storage repository.FileStorage, config *config.ServerConfig) *MetricsSaver {
	return &MetricsSaver{
		storage:  storage,
		config:   config,
		stopChan: make(chan struct{}),
	}
}

func (s *MetricsSaver) Run() error {
	// Загружаем метрики при старте, если включено
	if s.config.Restore {
		if err := s.storage.LoadFromFile(*s.config); err != nil {
			middlewares.Log.Error("Ошибка загрузки метрик: " + err.Error())
		}
	}

	// Запускаем периодическое сохранение
	go s.startMetricsSaver()

	return nil
}

func (s *MetricsSaver) Stop() error {
	close(s.stopChan)
	// Сохраняем метрики при остановке
	err := s.storage.SaveToFile(*s.config)
	if err != nil {
		middlewares.Log.Error("Ошибка сохранения метрик при остановке: " + err.Error())
	}
	return nil
}

func (s *MetricsSaver) startMetricsSaver() {
	// Если интервал 0, делаем синхронную запись
	if s.config.StoreInterval == 0 {
		return
	}

	ticker := time.NewTicker(s.config.StoreInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.storage.SaveToFile(*s.config); err != nil {
				middlewares.Log.Error("Ошибка сохранения метрик: " + err.Error())
			}
		case <-s.stopChan:
			return
		}
	}
}
