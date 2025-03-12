package metrics

import (
	"time"
)

// Примеры использования метрик в приложении

// TrackDatabaseQuery отслеживает время выполнения запроса к базе данных
func TrackDatabaseQuery(queryType string, f func() error) error {
	start := time.Now()
	err := f()
	duration := time.Since(start).Seconds()
	DatabaseQueryDuration.WithLabelValues(queryType).Observe(duration)
	return err
}

// TrackHTTPRequest отслеживает время выполнения HTTP-запроса
func TrackHTTPRequest(method, endpoint string, f func() (int, error)) (int, error) {
	start := time.Now()
	status, err := f()
	duration := time.Since(start).Seconds()
	HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
	HTTPRequestsTotal.WithLabelValues(method, endpoint, string(status)).Inc()
	return status, err
}

// IncrementParcelCreated увеличивает счетчик созданных посылок
func IncrementParcelCreated() {
	ParcelCreatedTotal.Inc()
}

// IncrementParcelStatusUpdated увеличивает счетчик обновлений статуса посылок
func IncrementParcelStatusUpdated(status string) {
	ParcelStatusUpdatedTotal.WithLabelValues(status).Inc()
}

// IncrementDeliveryCreated увеличивает счетчик созданных доставок
func IncrementDeliveryCreated() {
	DeliveryCreatedTotal.Inc()
}

// IncrementDeliveryStatusUpdated увеличивает счетчик обновлений статуса доставок
func IncrementDeliveryStatusUpdated(status string) {
	DeliveryStatusUpdatedTotal.WithLabelValues(status).Inc()
}

// IncrementPaymentProcessed увеличивает счетчик обработанных платежей
func IncrementPaymentProcessed(status, method string) {
	PaymentProcessedTotal.WithLabelValues(status, method).Inc()
}

// SetActiveConnections устанавливает количество активных WebSocket соединений
func SetActiveConnections(count float64) {
	ActiveConnections.Set(count)
}

// IncrementKafkaMessagesProcessed увеличивает счетчик обработанных сообщений Kafka
func IncrementKafkaMessagesProcessed(topic string) {
	KafkaMessagesProcessedTotal.WithLabelValues(topic).Inc()
}

// IncrementCacheHit увеличивает счетчик попаданий в кэш
func IncrementCacheHit() {
	CacheHitTotal.Inc()
}

// IncrementCacheMiss увеличивает счетчик промахов кэша
func IncrementCacheMiss() {
	CacheMissTotal.Inc()
}
