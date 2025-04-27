package handlers

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
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
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")
	metricValue := chi.URLParam(r, "value")

	// Проверяем что username не пустой
	if strings.TrimSpace(metricName) == "" || strings.ContainsAny(metricName, "!@#$%^&*") {
		render.Status(r, http.StatusNotFound)
		return
	}

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

/*
 * ValueHandler - обработчик информации о метрике
 */
func (s *MetricServer) ValueHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Разбираем URL: /value/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")

	// Проверяем что username не пустой
	if strings.TrimSpace(metricName) == "" || strings.ContainsAny(metricName, "!@#$%^&*") {
		render.Status(r, http.StatusNotFound)
		return
	}

	switch metricType {
	case "gauge":
		value, exist := s.storage.GetGauge(metricName)
		if !exist {
			http.Error(w, "Метрика не найдена", http.StatusNotFound)
			return
		}

		result := fmt.Sprintf("%f", value)
		render.Status(r, http.StatusOK)
		render.PlainText(w, r, result)
	case "counter":
		value, exist := s.storage.GetCounter(metricName)
		if !exist {
			http.Error(w, "Метрика не найдена", http.StatusNotFound)
			return
		}

		result := fmt.Sprintf("%f", value)
		render.Status(r, http.StatusOK)
		render.PlainText(w, r, result)
	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
	}
}

/*
 * MainHandler - обработчик информации о всех метрик
 */
func (s *MetricServer) MainHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	values := s.storage.GetAllGauges()

	render.Status(r, http.StatusOK)
	render.JSON(w, r, values)
}
