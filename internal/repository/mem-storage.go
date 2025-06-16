package repository

import (
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
