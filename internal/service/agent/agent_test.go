package agent

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"yupi/internal/config"
	"yupi/internal/httptransport/middlewares"
)

func TestNewAgent(t *testing.T) {
	serverURL := config.DefaultServerAddr
	pollInterval := config.DefaultPollInterval
	reportInterval := config.DefaultReportInterval
	useGzip := config.DefaultUseGzip

	agent := NewAgent(serverURL, pollInterval, reportInterval, useGzip)

	if agent == nil {
		t.Error("Не удалось инициировать агента")
	}

	if agent != nil && agent.serverURL != serverURL {
		t.Errorf("serverURL %v, ожидали %v", agent.serverURL, serverURL)
	}

	if agent != nil && agent.pollInterval != pollInterval {
		t.Errorf("Интервал обновления метрик %v, ожидали %v", agent.pollInterval, pollInterval)
	}

	if agent != nil && agent.reportInterval != reportInterval {
		t.Errorf("Интервал отправки метрик %v, ожидали %v", agent.reportInterval, reportInterval)
	}

	if agent != nil && agent.storage == nil {
		t.Error("Не удалось инициализировать хранилище")
	}
}

func TestAgent_SendMetric(t *testing.T) {
	// Создаем тестовый HTTP сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Ожидали тип запроса POST, получили %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	agent := NewAgent(server.URL, config.DefaultPollInterval, config.DefaultReportInterval, config.DefaultUseGzip)

	tests := []struct {
		name       string
		metricType string
		metricName string
		value      interface{}
		wantErr    bool
	}{
		{
			name:       "Успешное_обновление_метрики_с_типом_gauge",
			metricType: TypeGauge,
			metricName: "test_gauge",
			value:      42.0,
			wantErr:    false,
		},
		{
			name:       "Успешное_обновление_метрики_с_типом_counters",
			metricType: TypeCounter,
			metricName: "test_counter",
			value:      int64(10),
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := agent.sendMetric(tt.metricType, tt.metricName, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("sendMetric() ошибка %v, ожидали %v", err, tt.wantErr)
			}
		})
	}
}

func TestAgent_ReportMetrics(t *testing.T) {
	// Создаем тестовый HTTP сервер для подсчета отправленных метрик
	var receivedMetrics int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMetrics++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	agent := NewAgent(server.URL, config.DefaultPollInterval, config.DefaultReportInterval, config.DefaultUseGzip)

	// Добавляем тестовые метрики
	agent.storage.UpdateGauge("test_gauge1", 1.0)
	agent.storage.UpdateGauge("test_gauge2", 2.0)
	agent.storage.UpdateCounter(MetricCount, 1)

	// Отправляем PollCount
	count, exists := agent.storage.GetCounter(MetricCount)
	if exists {
		if err := agent.sendMetric(TypeCounter, MetricCount, count); err != nil {
			t.Error("Не удалось отправить метрику с типом counts")
		}
	}

	// Отправляем все gauge метрики
	for name, value := range agent.storage.GetAllGauges() {
		if err := agent.sendMetric(TypeGauge, name, value); err != nil {
			t.Error("Не удалось отправить метрику с типом gauge")
		}
	}
}

func TestAgent_SendMetricJSON_WithGzip(t *testing.T) {
	// Создаем тестовый сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем наличие заголовка Accept-Encoding
		if r.Header.Get("Accept-Encoding") != "gzip" {
			t.Error("Заголовок Accept-Encoding: gzip не установлен")
		}
		// Проверяем наличие заголовка Content-Encoding для запросов с включенным сжатием
		if r.Header.Get("Content-Encoding") == "gzip" {
			// Проверяем, что данные действительно сжаты
			reader, err := gzip.NewReader(r.Body)
			if err != nil {
				t.Error("Ошибка при чтении сжатых данных:", err)
				return
			}
			defer reader.Close()

			// Читаем и проверяем данные
			_, err = io.ReadAll(reader)
			if err != nil {
				t.Error("Ошибка при чтении разжатых данных:", err)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tests := []struct {
		name       string
		useGzip    bool
		metricType string
		metricName string
		value      interface{}
		wantErr    bool
	}{
		{
			name:       "Отправка_метрики_с_gzip_сжатием",
			useGzip:    true,
			metricType: TypeGauge,
			metricName: "test_gauge",
			value:      42.0,
			wantErr:    false,
		},
		{
			name:       "Отправка_метрики_без_gzip_сжатия",
			useGzip:    false,
			metricType: TypeGauge,
			metricName: "test_gauge",
			value:      42.0,
			wantErr:    false,
		},
		{
			name:       "Отправка_counter_метрики_с_gzip_сжатием",
			useGzip:    true,
			metricType: TypeCounter,
			metricName: "test_counter",
			value:      int64(10),
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := NewAgent(server.URL,
				config.DefaultPollInterval,
				config.DefaultReportInterval,
				tt.useGzip)

			err := agent.sendMetricJSON(tt.metricType, tt.metricName, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("sendMetricJSON() ошибка = %v, ожидали ошибку = %v", err, tt.wantErr)
			}
		})
	}
}

func TestGzipMiddleware(t *testing.T) {
	// Создаем тестовый обработчик
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test response"))
	})

	// Оборачиваем обработчик в middleware
	middleware := middlewares.GzipMiddleware(handler)

	tests := []struct {
		name            string
		acceptEncoding  string
		contentEncoding string
		wantCompressed  bool
	}{
		{
			name:            "Клиент_отправляет_сжатые_данные",
			acceptEncoding:  "",
			contentEncoding: "gzip",
			wantCompressed:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем тестовый запрос
			req := httptest.NewRequest("GET", "/", nil)
			if tt.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}
			if tt.contentEncoding != "" {
				req.Header.Set("Content-Encoding", tt.contentEncoding)
			}

			// Создаем тестовый ResponseRecorder
			rr := httptest.NewRecorder()

			// Выполняем запрос
			middleware.ServeHTTP(rr, req)

			// Проверяем результаты
			if tt.wantCompressed {
				if rr.Header().Get("Content-Encoding") != "gzip" {
					t.Error("Ответ должен быть сжат, но заголовок Content-Encoding: gzip отсутствует")
				}
			} else {
				if rr.Header().Get("Content-Encoding") == "gzip" {
					t.Error("Ответ не должен быть сжат, но заголовок Content-Encoding: gzip присутствует")
				}
			}
		})
	}
}
