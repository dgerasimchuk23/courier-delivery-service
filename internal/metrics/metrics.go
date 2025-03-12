package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTPRequestsTotal счетчик HTTP запросов
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// HTTPRequestDuration гистограмма времени ответа
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// ParcelCreatedTotal счетчик созданных посылок
	ParcelCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "parcel_created_total",
			Help: "Total number of created parcels",
		},
	)

	// ParcelStatusUpdatedTotal счетчик обновлений статуса посылок
	ParcelStatusUpdatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "parcel_status_updated_total",
			Help: "Total number of parcel status updates",
		},
		[]string{"status"},
	)

	// DeliveryCreatedTotal счетчик созданных доставок
	DeliveryCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "delivery_created_total",
			Help: "Total number of created deliveries",
		},
	)

	// DeliveryStatusUpdatedTotal счетчик обновлений статуса доставок
	DeliveryStatusUpdatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "delivery_status_updated_total",
			Help: "Total number of delivery status updates",
		},
		[]string{"status"},
	)

	// PaymentProcessedTotal счетчик обработанных платежей
	PaymentProcessedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "payment_processed_total",
			Help: "Total number of processed payments",
		},
		[]string{"status", "method"},
	)

	// ActiveConnections счетчик активных WebSocket соединений
	ActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "websocket_active_connections",
			Help: "Number of active WebSocket connections",
		},
	)

	// KafkaMessagesProcessedTotal счетчик обработанных сообщений Kafka
	KafkaMessagesProcessedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "kafka_messages_processed_total",
			Help: "Total number of Kafka messages processed",
		},
		[]string{"topic"},
	)

	// DatabaseQueryDuration гистограмма времени выполнения запросов к БД
	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "database_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"query_type"},
	)

	// CacheHitTotal счетчик попаданий в кэш
	CacheHitTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "cache_hit_total",
			Help: "Total number of cache hits",
		},
	)

	// CacheMissTotal счетчик промахов кэша
	CacheMissTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "cache_miss_total",
			Help: "Total number of cache misses",
		},
	)
)
