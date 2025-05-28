package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"io"
	"net/http"
	"strconv"
	"strings"
	"yupi/internal/domain/metrics"
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

func (s *MetricServer) JSONUpdateHandler(w http.ResponseWriter, r *http.Request) {
	var m metrics.Metrics
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	switch m.MType {
	case "gauge":
		s.storage.UpdateGaugeV2(&m)
	case "counter":
		s.storage.UpdateCounterV2(&m)
	default:
		http.Error(w, `{"error":"Invalid metric type"}`, http.StatusBadRequest)
		return
	}

	respondJSON(w, metrics.Metrics{ID: m.ID, MType: m.MType, Value: m.Value, Delta: m.Delta})
}

// UpdateHandler - обработчик обновления метрик
func (s *MetricServer) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()
	_, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
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

	defer r.Body.Close()

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

		result := fmt.Sprintf("%v", value)
		render.Status(r, http.StatusOK)
		render.PlainText(w, r, result)
	case "counter":
		value, exist := s.storage.GetCounter(metricName)
		if !exist {
			http.Error(w, "Метрика не найдена", http.StatusNotFound)
			return
		}

		result := fmt.Sprintf("%d", value)
		render.Status(r, http.StatusOK)
		render.PlainText(w, r, result)
	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
	}
}

func (s *MetricServer) JSONValueHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var m metrics.Metrics
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}

	switch m.MType {
	case "gauge":
		value, exist := s.storage.GetGauge(m.ID)
		if !exist {
			http.Error(w, `{"error":"Метрика не найдена"}`, http.StatusNotFound)
			return
		}
		m.Value = new(float64)
		*m.Value = value
	case "counter":
		delta, exist := s.storage.GetCounter(m.ID)
		if !exist {
			http.Error(w, "Метрика не найдена", http.StatusNotFound)
			return
		}
		m.Delta = new(int64)
		*m.Delta = delta
	default:
		http.Error(w, `{"error":"Invalid metric type"}`, http.StatusBadRequest)
		return
	}

	respondJSON(w, metrics.Metrics{ID: m.ID, MType: m.MType, Value: m.Value, Delta: m.Delta})
}

/*
 * MainHandler - обработчик информации о всех метрик
 */
func (s *MetricServer) MainHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()

	values := s.storage.GetAllGauges()
	// Устанавливаем Content-Type до сжатия
	w.Header().Set("Content-Type", "text/html")

	// Проверяем, поддерживает ли клиент gzip
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
	}

	render.Status(r, http.StatusOK)

	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(true)
	if err := enc.Encode(values); err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	w.Write(buf.Bytes()) //nolint:errcheck
}

func respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
	}
}
