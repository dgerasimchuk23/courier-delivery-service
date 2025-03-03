package cache

import (
	"context"
	"delivery/config"
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
	client *redis.Client
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

	log.Println("Успешное подключение к Redis")
	return &RedisClient{client: client}
}

// Set устанавливает значение по ключу с указанным временем жизни
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	log.Printf("Setting key %s in Redis with expiration %v", key, expiration)
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Get получает значение по ключу
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	log.Printf("Getting key %s from Redis", key)
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", ErrKeyNotFound
	}
	return val, err
}

// Delete удаляет ключ
func (r *RedisClient) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Close закрывает соединение с Redis
func (r *RedisClient) Close() error {
	return r.client.Close()
}
