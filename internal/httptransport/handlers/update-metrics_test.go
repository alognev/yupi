package handlers

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"yupi/internal/repository"
)

func TestMetricServer_UpdateHandler(t *testing.T) {
	storage := repository.NewMemStorage()
	server := NewMetricServer(storage)

	// Инициализация роутера
	r := chi.NewRouter()

	// Настройка маршрутов
	r.Post("/update/{type}/{name}/{value}", server.UpdateHandler)

	tests := []struct {
		name        string
		method      string
		path        string
		wantStatus  int
		wantGauge   float64
		wantCounter int64
		checkMetric bool
	}{
		{
			name:        "Успешное обновление метрики с типом gauge и значением с вещественным числом",
			method:      http.MethodPost,
			path:        "/update/gauge/test_gauge/42.0",
			wantStatus:  http.StatusOK,
			wantGauge:   42.0,
			checkMetric: true,
		},
		{
			name:        "Успешное обновление метрики с типом counters и значением с целым числом",
			method:      http.MethodPost,
			path:        "/update/counter/test_counter/10",
			wantStatus:  http.StatusOK,
			wantCounter: 10,
			checkMetric: true,
		},
		{
			name:        "Некорректный тип запроса",
			method:      http.MethodGet,
			path:        "/update/gauge/test/42.0",
			wantStatus:  http.StatusMethodNotAllowed,
			checkMetric: false,
		},
		{
			name:        "Некорректный формат URL",
			method:      http.MethodPost,
			path:        "/update/gauge",
			wantStatus:  http.StatusNotFound,
			checkMetric: false,
		},
		{
			name:        "Некорректное значение метрики с типом gauge",
			method:      http.MethodPost,
			path:        "/update/gauge/test/invalid",
			wantStatus:  http.StatusBadRequest,
			checkMetric: false,
		},
		{
			name:        "Некорректное значение метрики с типом counter",
			method:      http.MethodPost,
			path:        "/update/counter/test/invalid",
			wantStatus:  http.StatusBadRequest,
			checkMetric: false,
		},
		{
			name:        "Некорректное тип метрики",
			method:      http.MethodPost,
			path:        "/update/invalid/test/42",
			wantStatus:  http.StatusBadRequest,
			checkMetric: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			//w := httptest.NewRecorder()

			//server.UpdateHandler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("UpdateHandler() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.checkMetric {
				if strings.Contains(tt.path, "gauge") {
					if value, exists := storage.GetGauge("test_gauge"); !exists || value != tt.wantGauge {
						t.Errorf("UpdateHandler() gauge = %v, want %v", value, tt.wantGauge)
					}
				} else if strings.Contains(tt.path, "counter") {
					metricName := strings.Split(tt.path, "/")[3]
					if value, exists := storage.GetCounter(metricName); !exists || value != tt.wantCounter {
						t.Errorf("UpdateHandler() counter = %v, want %v", value, tt.wantCounter)
					}
				}
			}
		})
	}
}

func TestNewMetricServer(t *testing.T) {
	storage := repository.NewMemStorage()
	server := NewMetricServer(storage)

	if server == nil {
		t.Error("Не удалось инициализировать сервер метрик")
	}

	if server.storage != storage {
		t.Error("Хранилище сервиса метрик имеет другой тип")
	}
}
