package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"strings"
	"time"
	"yupi/internal/domain/metrics"
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
	useGzip        bool
}

// Конструктор
func NewAgent(serverURL string, pollInterval int64, reportInterval int64, useGzip bool) *Agent {

	return &Agent{
		protocol:       "http",
		serverURL:      serverURL,
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		storage:        repository.NewMemStorage(),
		useGzip:        useGzip,
	}
}

func (a *Agent) Run() {
	pollTicker := time.NewTicker(time.Duration(a.pollInterval) * time.Second)
	reportTicker := time.NewTicker(time.Duration(a.reportInterval) * time.Second)
	defer func() {
		pollTicker.Stop()
		reportTicker.Stop()
	}()

	for {
		select {
		case <-pollTicker.C:
			a.aggregateMetrics()
		case <-reportTicker.C:
			if err := a.reportMetrics(); err != nil {
				fmt.Printf("Error reporting metrics: %v\n", err)
			}
		}
	}
}

// Агрегирование метрик
func (a *Agent) aggregateMetrics() {
	var stats runtime.MemStats

	runtime.ReadMemStats(&stats)
	metricList := map[string]interface{}{
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

	for k, v := range metricList {
		a.storage.UpdateGauge(k, v.(float64))
	}

	a.storage.UpdateCounter("PollCount", 1)
}

// Отправка всех метрик на сервер
func (a *Agent) reportMetrics() error {

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

	return nil
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
	url := fmt.Sprintf("%s/%s/", a.serverURL, UpdateURL)

	// Добавляем http://, если URL не начинается с протокола
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = a.protocol + "://" + url
	}

	// Создаем структуру метрики
	metric := metrics.Metrics{
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

	// Подготавливаем тело запроса (сжатое или обычное)
	var body bytes.Buffer
	if a.useGzip {
		gz := gzip.NewWriter(&body)
		if _, err := gz.Write(jsonData); err != nil {
			return fmt.Errorf("failed to compress data: %w", err)
		}
		if err := gz.Close(); err != nil {
			return fmt.Errorf("failed to close gzip writer: %w", err)
		}
	} else {
		body.Write(jsonData)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest(
		"POST",
		url,
		&body,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Устанавливаем заголовки
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")

	if a.useGzip {
		req.Header.Set("Content-Encoding", "gzip")
	}

	resp, err := client.Do(req)

	// Отправляем запрос
	//resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))

	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Обрабатываем возможный сжатый ответ
	var reader io.Reader
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	default:
		reader = resp.Body
	}

	// Читаем ответ сервера (для логирования или проверки)
	responseBody, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %d, body: %s",
			resp.StatusCode, string(responseBody))
	}

	return nil
}
