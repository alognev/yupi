package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"yupi/internal/config"
)

type FileStorage struct {
	memStorage *MemStorage
}

func NewFileStorage(storage *MemStorage) *FileStorage {
	return &FileStorage{
		memStorage: storage,
	}
}

func (fs *FileStorage) SaveToFile(config config.ServerConfig) error {
	data := StorageData{
		Gauges:   fs.memStorage.GetAllGauges(),
		Counters: make(map[string]int64),
	}

	// Получаем все counters (нужно добавить метод GetAllCounters в MemStorage)
	fs.memStorage.mu.RLock()
	for k, v := range fs.memStorage.counters {
		data.Counters[k] = v
	}
	fs.memStorage.mu.RUnlock()

	fileData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Создаем директорию если нужно
	if err := os.MkdirAll(filepath.Dir(config.FileStoragePath), 0755); err != nil {
		return fmt.Errorf("не удалось создать директорию: %w", err)
	}

	return os.WriteFile(config.FileStoragePath, fileData, 0644)
}

func (fs *FileStorage) LoadFromFile(config config.ServerConfig) error {
	fileData, err := os.ReadFile(config.FileStoragePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Если файл не существует, это не ошибка при первом запуске
			return nil
		}
		return fmt.Errorf("ошибка чтения файла %s: %w", config.FileStoragePath, err)
	}

	var data StorageData
	if err := json.Unmarshal(fileData, &data); err != nil {
		return err
	}

	// Обновляем данные в хранилище
	for name, value := range data.Gauges {
		fs.memStorage.UpdateGauge(name, value)
	}

	for name, value := range data.Counters {
		fs.memStorage.UpdateCounter(name, value)
	}

	return nil
}
