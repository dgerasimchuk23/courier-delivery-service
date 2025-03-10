package cache

import (
	"context"
	"delivery/internal/config"
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/redis/go-redis/v9"
)

// Ошибки
var (
	ErrKeyNotFound = errors.New("ключ не найден")
)

// RedisClient представляет клиент Redis
type RedisClient struct {
	client       *redis.Client
	cleanupDone  chan bool
	cleanupStats *RedisStats
}

// Проверка, что RedisClient реализует интерфейс RedisClientInterface
var _ RedisClientInterface = (*RedisClient)(nil)

// NewRedisClient создает новый клиент Redis
func NewRedisClient(config *config.Config) *RedisClient {
	// Проверяем, доступен ли Redis
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", config.Redis.Host, config.Redis.Port), 2*time.Second)
	if err != nil {
		log.Printf("Redis недоступен по адресу %s:%d: %v", config.Redis.Host, config.Redis.Port, err)
		return nil
	}
	if conn != nil {
		conn.Close()
	}

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Redis.Host, config.Redis.Port),
		Password: config.Redis.Password,
		DB:       config.Redis.DB,
	})

	// Проверка соединения
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("Ошибка подключения к Redis: %v", err)
		return nil
	}

	redisClient := &RedisClient{
		client: client,
	}

	// Запускаем периодическую очистку Redis
	redisClient.cleanupDone = redisClient.ScheduleRedisCleanup(1 * time.Hour)

	// Получаем начальную статистику
	stats, err := redisClient.MonitorStats(ctx)
	if err != nil {
		log.Printf("Ошибка при получении статистики Redis: %v", err)
	} else {
		redisClient.cleanupStats = stats
		log.Printf("Начальная статистика Redis: total keys=%d, rate limit keys=%d, blacklist keys=%d",
			stats.TotalKeys, stats.RateLimitKeys, stats.BlacklistKeys)
	}

	log.Println("Успешное подключение к Redis")
	return redisClient
}

// Set устанавливает значение по ключу с указанным временем жизни
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if r == nil || r.client == nil {
		return fmt.Errorf("Redis client is nil")
	}

	log.Printf("Setting key %s in Redis with expiration %v", key, expiration)
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Get получает значение по ключу
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	if r == nil || r.client == nil {
		return "", fmt.Errorf("Redis client is nil")
	}

	log.Printf("Getting key %s from Redis", key)
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", ErrKeyNotFound
	}
	return val, err
}

// Delete удаляет ключ
func (r *RedisClient) Delete(ctx context.Context, key string) error {
	if r == nil || r.client == nil {
		return fmt.Errorf("Redis client is nil")
	}

	log.Printf("Deleting key %s from Redis", key)
	return r.client.Del(ctx, key).Err()
}

// Close закрывает соединение с Redis
func (r *RedisClient) Close() error {
	if r == nil || r.client == nil {
		return nil
	}

	// Останавливаем периодическую очистку
	if r.cleanupDone != nil {
		r.cleanupDone <- true
	}

	// Получаем финальную статистику
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stats, err := r.MonitorStats(ctx)
	if err != nil {
		log.Printf("Ошибка при получении финальной статистики Redis: %v", err)
	} else if r.cleanupStats != nil {
		log.Printf("Финальная статистика Redis: total keys=%d, rate limit keys=%d, blacklist keys=%d",
			stats.TotalKeys, stats.RateLimitKeys, stats.BlacklistKeys)
		log.Printf("Изменение статистики Redis: total keys=%d, rate limit keys=%d, blacklist keys=%d",
			stats.TotalKeys-r.cleanupStats.TotalKeys,
			stats.RateLimitKeys-r.cleanupStats.RateLimitKeys,
			stats.BlacklistKeys-r.cleanupStats.BlacklistKeys)
	}

	log.Println("Закрытие соединения с Redis")
	return r.client.Close()
}
