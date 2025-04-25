package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"yupi/internal/repository"
)

// MetricServer - сервер для обработки метрик
type MetricServer struct {
	storage *repository.MemStorage
}

// NewMetricServer - конструктор сервера метрик
func NewMetricServer(storage *repository.MemStorage) *MetricServer {
	return &MetricServer{storage: storage}
}

// UpdateHandler - обработчик обновления метрик
func (s *MetricServer) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Разбираем URL: /update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 5 {
		http.Error(w, "Invalid URL format", http.StatusNotFound)
		return
	}

	metricType := parts[2]
	metricName := parts[3]
	metricValue := parts[4]

	switch metricType {
	case "gauge":
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(w, "Invalid gauge value", http.StatusBadRequest)
			return
		}
		s.storage.UpdateGauge(metricName, value)
		w.WriteHeader(http.StatusOK)
	case "counter":
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(w, "Invalid counter value", http.StatusBadRequest)
			return
		}
		s.storage.UpdateCounter(metricName, value)
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
	}
}
