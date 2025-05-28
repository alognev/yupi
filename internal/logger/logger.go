package logger

import (
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"net/http"
	"time"
)

// Log будет доступен всему коду как синглтон.
// Никакой код навыка, кроме функции Initialize, не должен модифицировать эту переменную.
// По умолчанию установлен no-op-логер, который не выводит никаких сообщений.
var Log *zap.Logger = zap.NewNop()

// Initialize инициализирует синглтон логера с необходимым уровнем логирования.
func Initialize(level string) error {
	// преобразуем текстовый уровень логирования в zap.AtomicLevel
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	// создаём новую конфигурацию логера
	cfg := zap.NewProductionConfig()
	// устанавливаем уровень
	cfg.Level = lvl
	// создаём логер на основе конфигурации
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	// устанавливаем синглтон
	Log = zl
	return nil
}

// LoggingMiddleware - middleware-логер для входящих HTTP-запросов
func LoggingRequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		//body, _ := io.ReadAll(r.Body)
		//r.Body = io.NopCloser(bytes.NewBuffer(body))
		// Обертка для получения статуса ответа
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		// Передаем запрос следующему обработчику
		next.ServeHTTP(ww, r)
		//defer r.Body.Close()
		// Логируем информацию после обработки запроса
		duration := time.Since(start)
		Log.Info("got incoming HTTP request",
			zap.String("method", r.Method),
			zap.String("URI", r.URL.Path),
			//zap.String("body", string(body)),
			zap.Duration("duration", duration.Round(time.Millisecond)),
			zap.Int("status", ww.Status()),
			zap.Int("bytes", ww.BytesWritten()),
		)
	})
}
