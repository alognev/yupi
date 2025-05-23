package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"yupi/internal/repository"

	"github.com/go-chi/chi/v5"
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
			name:        "Успешное_обновление_метрики_с_типом_gauge_и_значением_с_вещественным_числом",
			method:      http.MethodPost,
			path:        "/update/gauge/test_gauge/42.0",
			wantStatus:  http.StatusOK,
			wantGauge:   42.0,
			checkMetric: true,
		},
		{
			name:        "Успешное_обновление_метрики_с_типом_counters_и_значением_с_целым_числом",
			method:      http.MethodPost,
			path:        "/update/counter/test_counter/10",
			wantStatus:  http.StatusOK,
			wantCounter: 10,
			checkMetric: true,
		},
		{
			name:        "Некорректный_тип_запроса",
			method:      http.MethodGet,
			path:        "/update/gauge/test/42.0",
			wantStatus:  http.StatusMethodNotAllowed,
			checkMetric: false,
		},
		{
			name:        "Некорректный_формат_URL",
			method:      http.MethodPost,
			path:        "/update/gauge",
			wantStatus:  http.StatusNotFound,
			checkMetric: false,
		},
		{
			name:        "Некорректное_значение_метрики_с_типом_gauge",
			method:      http.MethodPost,
			path:        "/update/gauge/test/invalid",
			wantStatus:  http.StatusBadRequest,
			checkMetric: false,
		},
		{
			name:        "Некорректное_значение_метрики_с_типом_counter",
			method:      http.MethodPost,
			path:        "/update/counter/test/invalid",
			wantStatus:  http.StatusBadRequest,
			checkMetric: false,
		},
		{
			name:        "Некорректное_тип_метрики",
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

	if server != nil && server.storage != storage {
		t.Error("Хранилище сервиса метрик имеет другой тип")
	}
}

func TestMetricServer_JSONUpdateHandler(t *testing.T) {
	storage := repository.NewMemStorage()
	server := NewMetricServer(storage)

	tests := []struct {
		name             string
		requestBody      string
		wantStatus       int
		wantGaugeValue   float64
		wantCounterValue int64
		checkMetric      bool
	}{
		{
			name:           "Успешное_обновление_метрики_с_типом_gauge",
			requestBody:    `{"id":"test_gauge","type":"gauge","value":42.5}`,
			wantStatus:     http.StatusOK,
			wantGaugeValue: 42.5,
			checkMetric:    true,
		},
		{
			name:             "Успешное_обновление_метрики_с_типом_counter",
			requestBody:      `{"id":"test_counter","type":"counter","delta":10}`,
			wantStatus:       http.StatusOK,
			wantCounterValue: 10,
			checkMetric:      true,
		},
		{
			name:        "Некорректный_JSON",
			requestBody: `{"id":"test","type":"gauge","value":42.5`,
			wantStatus:  http.StatusBadRequest,
			checkMetric: false,
		},
		{
			name:        "Некорректный_тип_метрики",
			requestBody: `{"id":"test","type":"invalid","value":42.5}`,
			wantStatus:  http.StatusBadRequest,
			checkMetric: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/update", strings.NewReader(tt.requestBody))
			w := httptest.NewRecorder()

			server.JSONUpdateHandler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("JSONUpdateHandler() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.checkMetric {
				if strings.Contains(tt.requestBody, "gauge") {
					if value, exists := storage.GetGauge("test_gauge"); !exists || value != tt.wantGaugeValue {
						t.Errorf("JSONUpdateHandler() gauge = %v, want %v", value, tt.wantGaugeValue)
					}
				} else if strings.Contains(tt.requestBody, "counter") {
					if value, exists := storage.GetCounter("test_counter"); !exists || value != tt.wantCounterValue {
						t.Errorf("JSONUpdateHandler() counter = %v, want %v", value, tt.wantCounterValue)
					}
				}
			}
		})
	}
}

func TestMetricServer_JSONValueHandler(t *testing.T) {
	storage := repository.NewMemStorage()
	server := NewMetricServer(storage)

	reqBody := `{"id":"test_gauge","type":"gauge","value":42.5}`
	req := httptest.NewRequest(http.MethodPost, "/update", strings.NewReader(reqBody))
	w := httptest.NewRecorder()
	server.JSONUpdateHandler(w, req)

	req2Body := `{"id":"test_counter","type":"counter","delta":10}`
	req2 := httptest.NewRequest(http.MethodPost, "/update", strings.NewReader(req2Body))
	w2 := httptest.NewRecorder()
	server.JSONUpdateHandler(w2, req2)

	tests := []struct {
		name        string
		requestBody string
		wantStatus  int
		wantBody    string
	}{
		{
			name:        "Успешное_получение_метрики_с_типом_gauge",
			requestBody: `{"id":"test_gauge","type":"gauge"}`,
			wantStatus:  http.StatusOK,
			wantBody:    `{"id":"test_gauge","type":"gauge","value":42.5}`,
		},
		{
			name:        "Успешное_получение_метрики_с_типом_counter",
			requestBody: `{"id":"test_counter","type":"counter"}`,
			wantStatus:  http.StatusOK,
			wantBody:    `{"id":"test_counter","type":"counter","delta":10}`,
		},
		{
			name:        "Некорректный_JSON",
			requestBody: `{"id":"test","type":"gauge"`,
			wantStatus:  http.StatusBadRequest,
			wantBody:    `{"error":"invalid JSON"}`,
		},
		{
			name:        "Некорректный_тип_метрики",
			requestBody: `{"id":"test","type":"invalid"}`,
			wantStatus:  http.StatusBadRequest,
			wantBody:    `{"error":"Invalid metric type"}`,
		},
		{
			name:        "Метрика_не_найдена",
			requestBody: `{"id":"nonexistent","type":"gauge"}`,
			wantStatus:  http.StatusNotFound,
			wantBody:    `{"error":"Метрика не найдена"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/value", strings.NewReader(tt.requestBody))
			w := httptest.NewRecorder()

			server.JSONValueHandler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("JSONValueHandler() status = %v, want %v", w.Code, tt.wantStatus)
			}

			// Проверяем тело ответа
			gotBody := strings.TrimSpace(w.Body.String())
			if gotBody != tt.wantBody {
				t.Errorf("JSONValueHandler() body = %v, want %v", gotBody, tt.wantBody)
			}
		})
	}
}
