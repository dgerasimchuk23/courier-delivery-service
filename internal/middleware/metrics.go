package middleware

import (
	"delivery/internal/metrics"
	"net/http"
	"strconv"
	"time"
)

// MetricsMiddleware middleware для сбора метрик HTTP-запросов
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Обертка для записи статуса ответа
		ww := NewResponseWriter(w)

		// Вызов следующего обработчика
		next.ServeHTTP(ww, r)

		// Запись метрик
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(ww.Status())

		metrics.HTTPRequestsTotal.WithLabelValues(r.Method, r.URL.Path, status).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
	})
}

// ResponseWriter обертка для http.ResponseWriter для отслеживания статуса ответа
type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// NewResponseWriter создает новый ResponseWriter
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{w, http.StatusOK}
}

// WriteHeader переопределяет метод WriteHeader для сохранения статуса ответа
func (rw *ResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Status возвращает статус ответа
func (rw *ResponseWriter) Status() int {
	return rw.statusCode
}
