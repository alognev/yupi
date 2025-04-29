package agent

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"yupi/internal/config"
)

func TestNewAgent(t *testing.T) {
	serverURL := config.DefaultServerAddr
	pollInterval := config.DefaultPollInterval
	reportInterval := config.DefaultReportInterval

	agent := NewAgent(serverURL, pollInterval, reportInterval)

	if agent == nil {
		t.Error("Не удалось инициировать агента")
	}

	if agent.serverURL != serverURL {
		t.Errorf("serverURL %v, ожидали %v", agent.serverURL, serverURL)
	}

	if agent.pollInterval != pollInterval {
		t.Errorf("Интервал обновления метрик %v, ожидали %v", agent.pollInterval, pollInterval)
	}

	if agent.reportInterval != reportInterval {
		t.Errorf("Интервал отправки метрик %v, ожидали %v", agent.reportInterval, reportInterval)
	}

	if agent.storage == nil {
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

	agent := NewAgent(server.URL, config.DefaultPollInterval, config.DefaultReportInterval)

	tests := []struct {
		name       string
		metricType string
		metricName string
		value      interface{}
		wantErr    bool
	}{
		{
			name:       "Успешное обновление метрики с типом gauge",
			metricType: TypeGauge,
			metricName: "test_gauge",
			value:      42.0,
			wantErr:    false,
		},
		{
			name:       "Успешное обновление метрики с типом counters",
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

	agent := NewAgent(server.URL, config.DefaultPollInterval, config.DefaultReportInterval)

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
