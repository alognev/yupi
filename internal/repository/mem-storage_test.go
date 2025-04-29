package repository

import (
	"testing"
)

func TestNewMemStorage(t *testing.T) {
	storage := NewMemStorage()

	if storage == nil {
		t.Error("Не удалось инициализировать хранилище")
	}

	if storage == nil || storage.gauges == nil {
		t.Error("Не удалось инициализировать хранилищи метрик с типом gauges")
	}

	if storage == nil || storage.counters == nil {
		t.Error("Не удалось инициализировать хранилищи метрик с типом counters")
	}
}

func TestMemStorage_UpdateGauge(t *testing.T) {
	storage := NewMemStorage()

	tests := []struct {
		name  string
		value float64
	}{
		{"test1", 1.0},
		{"test2", -1.0},
		{"test3", 0.0},
		{"test4", 999999.999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage.UpdateGauge(tt.name, tt.value)

			got, exists := storage.GetGauge(tt.name)
			if !exists {
				t.Errorf("Не удалось записать метрику %s с типом gauge", tt.name)
			}

			if got != tt.value {
				t.Errorf("Получили %v при ожидаемом результате %v", got, tt.value)
			}
		})
	}
}

func TestMemStorage_UpdateCounter(t *testing.T) {
	storage := NewMemStorage()

	tests := []struct {
		name     string
		values   []int64
		expected int64
	}{
		{
			name:     "Успешная запись одного значения",
			values:   []int64{5},
			expected: 5,
		},
		{
			name:     "Успешная запись нескольких значений",
			values:   []int64{1, 2, 3},
			expected: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, v := range tt.values {
				storage.UpdateCounter(tt.name, v)
			}

			got, exists := storage.GetCounter(tt.name)
			if !exists {
				t.Errorf("Не удалось записать метрику %s с типом counter", tt.name)
			}

			if got != tt.expected {
				t.Errorf("Получили %v при ожидаемом результате %v", got, tt.expected)
			}
		})
	}
}

func TestMemStorage_GetAllGauges(t *testing.T) {
	storage := NewMemStorage()

	testData := map[string]float64{
		"gauge1": 1.0,
		"gauge2": 2.0,
		"gauge3": 3.0,
	}

	// Заполняем хранилище тестовыми данными
	for name, value := range testData {
		storage.UpdateGauge(name, value)
	}

	// Получаем все метрики
	gauges := storage.GetAllGauges()

	// Проверяем количество метрик
	if len(gauges) != len(testData) {
		t.Errorf("Получили кол-во записей метрик с типом gauges и значением %d , ожидали %d", len(gauges), len(testData))
	}

	// Проверяем значения
	for name, expectedValue := range testData {
		if actualValue, exists := gauges[name]; !exists {
			t.Errorf("Не удалось найти метрику %s с типом gauge", name)
		} else if actualValue != expectedValue {
			t.Errorf("Получили значение метрики %s с типом gauge %v, ожидали %v", name, actualValue, expectedValue)
		}
	}
}
