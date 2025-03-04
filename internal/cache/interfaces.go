package cache

import (
	"context"
	"time"
)

// RedisClientInterface определяет интерфейс для работы с Redis
type RedisClientInterface interface {
	// Set устанавливает значение по ключу с указанным временем жизни
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error

	// Get получает значение по ключу
	Get(ctx context.Context, key string) (string, error)

	// Delete удаляет ключ
	Delete(ctx context.Context, key string) error

	// Close закрывает соединение с Redis
	Close() error

	// SetJSON сохраняет объект в кэше в формате JSON
	SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error

	// GetJSON получает объект из кэша и десериализует его из JSON
	GetJSON(ctx context.Context, key string, dest interface{}) error

	// MonitorStats получает статистику использования Redis
	MonitorStats(ctx context.Context) (*RedisStats, error)

	// CleanupRateLimitKeys удаляет устаревшие ключи rate limit
	CleanupRateLimitKeys(ctx context.Context) (int64, error)

	// CleanupBlacklistKeys удаляет устаревшие ключи из черного списка токенов
	CleanupBlacklistKeys(ctx context.Context) (int64, error)

	// ScheduleRedisCleanup запускает периодическую очистку Redis
	ScheduleRedisCleanup(interval time.Duration) chan bool
}
