package agent

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"
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
	pollInterval   time.Duration
	reportInterval time.Duration
	storage        *repository.MemStorage
}

// Конструктор
func NewAgent(serverURL string, pollInterval time.Duration, reportInterval time.Duration) *Agent {

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
	runtime.ReadMemStats(&stats)

	for {

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
		time.Sleep(a.pollInterval)
	}
}

// Отправка всех метрик на сервер
func (a *Agent) reportMetrics(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		// Отправляем PollCount
		count, exists := a.storage.GetCounter(MetricCount)
		if exists {
			if err := a.sendMetric(TypeCounter, MetricCount, count); err != nil {
				return
			}
		}

		// Отправляем все gauge метрики
		for name, value := range a.storage.GetAllGauges() {
			if err := a.sendMetric(TypeGauge, name, value); err != nil {
				return
			}
		}

		time.Sleep(a.reportInterval)
	}
}

// Отправка метрики ан сервер
func (a *Agent) sendMetric(metricType, metricName string, value interface{}) error {
	url := fmt.Sprintf("%s/%s/%s/%s/%v", a.serverURL, UpdateURL, metricType, metricName, value)

	// Добавляем http://, если URL не начинается с протокола, столкнулся с проблемой, что в тестах он добавляется, а в агенте нет из-за чего виснет запрос
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = a.protocol + "://" + url
	}

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
