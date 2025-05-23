package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"
	. "yupi/internal/domain/metrics"
	"yupi/internal/repository"
)

const (
	TypeGauge   = "gauge"
	TypeCounter = "counter"
	MetricCount = "PollCount"
	UpdateURL   = "update"
)

type Agent struct {
	protocol       string
	serverURL      string
	pollInterval   int64
	reportInterval int64
	storage        *repository.MemStorage
}

// Конструктор
func NewAgent(serverURL string, pollInterval int64, reportInterval int64) *Agent {

	return &Agent{
		protocol:       "http",
		serverURL:      serverURL,
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		storage:        repository.NewMemStorage(),
	}
}

// Запуск сервиса по работе с метриками
func (a *Agent) Run() {
	var wg sync.WaitGroup
	wg.Add(2)

	go a.aggregateMetrics(&wg)
	go a.reportMetrics(&wg)

	wg.Wait()
}

// Агрегирование метрик
func (a *Agent) aggregateMetrics(wg *sync.WaitGroup) {
	defer wg.Done()
	var stats runtime.MemStats

	for {
		runtime.ReadMemStats(&stats)
		metrics := map[string]interface{}{
			"Alloc":         float64(stats.Alloc),
			"BuckHashSys":   float64(stats.BuckHashSys),
			"Frees":         float64(stats.Frees),
			"GCCPUFraction": stats.GCCPUFraction,
			"GCSys":         float64(stats.GCSys),
			"HeapAlloc":     float64(stats.HeapAlloc),
			"HeapIdle":      float64(stats.HeapIdle),
			"HeapInuse":     float64(stats.HeapInuse),
			"HeapObjects":   float64(stats.HeapObjects),
			"HeapReleased":  float64(stats.HeapReleased),
			"HeapSys":       float64(stats.HeapSys),
			"LastGC":        float64(stats.LastGC),
			"Lookups":       float64(stats.Lookups),
			"MCacheInuse":   float64(stats.MCacheInuse),
			"MCacheSys":     float64(stats.MCacheSys),
			"MSpanInuse":    float64(stats.MSpanInuse),
			"MSpanSys":      float64(stats.MSpanSys),
			"Mallocs":       float64(stats.Mallocs),
			"NextGC":        float64(stats.NextGC),
			"NumForcedGC":   float64(stats.NumForcedGC),
			"NumGC":         float64(stats.NumGC),
			"OtherSys":      float64(stats.OtherSys),
			"PauseTotalNs":  float64(stats.PauseTotalNs),
			"StackInuse":    float64(stats.StackInuse),
			"StackSys":      float64(stats.StackSys),
			"Sys":           float64(stats.Sys),
			"TotalAlloc":    float64(stats.TotalAlloc),
			"RandomValue":   rand.Float64(),
		}

		for k, v := range metrics {
			a.storage.UpdateGauge(k, v.(float64))
		}

		a.storage.UpdateCounter("PollCount", 1)
		time.Sleep(time.Duration(a.pollInterval) * time.Second)
	}
}

// Отправка всех метрик на сервер
func (a *Agent) reportMetrics(wg *sync.WaitGroup) error {
	defer wg.Done()
	for {
		// ждем время сбора метрик
		time.Sleep(time.Duration(a.reportInterval) * time.Second)
		// Отправляем PollCount
		count, exists := a.storage.GetCounter(MetricCount)
		if exists {
			if err := a.sendMetricJSON(TypeCounter, MetricCount, count); err != nil {
				log.Println(err)
				return err
			}
		}

		// Отправляем все gauge метрики
		for name, value := range a.storage.GetAllGauges() {
			if err := a.sendMetricJSON(TypeGauge, name, value); err != nil {
				log.Println(err)
				return err
			}
		}
	}
}

// Отправка метрики на сервер
func (a *Agent) sendMetric(metricType, metricName string, value interface{}) error {
	url := fmt.Sprintf("%s/%s/%s/%s/%v", a.serverURL, UpdateURL, metricType, metricName, value)

	// Добавляем http://, если URL не начинается с протокола, столкнулся с проблемой, что в тестах он добавляется, а в агенте нет из-за чего виснет запрос
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = a.protocol + "://" + url
	}
	log.Println(url)
	resp, err := http.Post(url, "text/plain", bytes.NewBufferString(""))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %d", resp.StatusCode)
	}
	return nil
}

// Отправка метрики на сервер в формате JSON
func (a *Agent) sendMetricJSON(metricType, metricName string, value interface{}) error {
	url := fmt.Sprintf("%s/update/", a.serverURL)

	// Добавляем http://, если URL не начинается с протокола
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = a.protocol + "://" + url
	}

	// Создаем структуру метрики
	metric := Metrics{
		ID:    metricName,
		MType: metricType,
	}

	// Заполняем значение в зависимости от типа метрики
	switch metricType {
	case TypeGauge:
		if val, ok := value.(float64); ok {
			metric.Value = &val
		} else {
			return fmt.Errorf("invalid gauge value type: %T", value)
		}
	case TypeCounter:
		if val, ok := value.(int64); ok {
			metric.Delta = &val
		} else {
			return fmt.Errorf("invalid counter value type: %T", value)
		}
	default:
		return fmt.Errorf("invalid metric type: %s", metricType)
	}

	// Сериализуем метрику в JSON
	jsonData, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("failed to marshal metric: %w", err)
	}

	// Отправляем запрос
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %d", resp.StatusCode)
	}
	return nil
}
