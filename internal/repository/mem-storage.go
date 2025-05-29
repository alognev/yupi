package repository

import (
	"encoding/json"
	"os"
	"sync"
	"yupi/internal/domain/metrics"
)

// MemStorage - хранилище метрик в памяти
type MemStorage struct {
	mu       sync.RWMutex
	gauges   map[string]float64
	counters map[string]int64
}

// NewMemStorage - конструктор хранилища
func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

// UpdateGauge - обновление метрики типа gauge
func (s *MemStorage) UpdateGauge(name string, value float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gauges[name] = value
}

func (s *MemStorage) UpdateGaugeV2(m *metrics.Metrics) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gauges[m.ID] = *m.Value
}

// UpdateCounter - обновление метрики типа counter
func (s *MemStorage) UpdateCounterV2(m *metrics.Metrics) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counters[m.ID] += *m.Delta
}

// UpdateCounter - обновление метрики типа counter
func (s *MemStorage) UpdateCounter(name string, value int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counters[name] += value
}

// GetGauge - получение значения gauge
func (s *MemStorage) GetGauge(name string) (float64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.gauges[name]
	return val, ok
}

// GetAllGauges возвращает все gauge-метрики из хранилища
func (s *MemStorage) GetAllGauges() map[string]float64 {
	s.mu.RLock()         // Блокируем для чтения
	defer s.mu.RUnlock() // Гарантируем разблокировку

	// Создаем копию для безопасного доступа (так подсказал deepseek %))
	gaugesCopy := make(map[string]float64, len(s.gauges))
	for k, v := range s.gauges {
		gaugesCopy[k] = v
	}

	return gaugesCopy
}

// GetCounter - получение значения counter
func (s *MemStorage) GetCounter(name string) (int64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.counters[name]
	return val, ok
}

// StorageData структура для сериализации данных
type StorageData struct {
	Gauges   map[string]float64 `json:"gauges"`
	Counters map[string]int64   `json:"counters"`
}

// SaveToFile сохраняет текущее состояние метрик в файл
func (s *MemStorage) SaveToFile(filepath string) error {
	s.mu.RLock()
	data := StorageData{
		Gauges:   make(map[string]float64),
		Counters: make(map[string]int64),
	}

	// Копируем данные
	for k, v := range s.gauges {
		data.Gauges[k] = v
	}
	for k, v := range s.counters {
		data.Counters[k] = v
	}
	s.mu.RUnlock()

	// Сериализуем данные
	fileData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Записываем в файл
	return os.WriteFile(filepath, fileData, 0644)
}

// LoadFromFile загружает состояние метрик из файла
func (s *MemStorage) LoadFromFile(filepath string) error {
	fileData, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	var data StorageData
	if err := json.Unmarshal(fileData, &data); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Загружаем данные
	for k, v := range data.Gauges {
		s.gauges[k] = v
	}
	for k, v := range data.Counters {
		s.counters[k] = v
	}

	return nil
}
