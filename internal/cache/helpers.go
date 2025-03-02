package cache

import (
	"context"
	"encoding/json"
	"log"
	"time"
)

// CacheItem представляет элемент кэша
type CacheItem struct {
	Value      interface{}
	Expiration time.Duration
}

// SetJSON сохраняет объект в кэше в формате JSON
func (r *RedisClient) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return r.Set(ctx, key, data, expiration)
}

// GetJSON получает объект из кэша и десериализует его из JSON
func (r *RedisClient) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := r.Get(ctx, key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), dest)
}

// GetOrSet получает значение из кэша или вызывает функцию для получения значения и сохранения в кэше
func (r *RedisClient) GetOrSet(ctx context.Context, key string, expiration time.Duration, fn func() (interface{}, error)) (string, error) {
	// Пытаемся получить из кэша
	val, err := r.Get(ctx, key)
	if err == nil {
		// Значение найдено в кэше
		return val, nil
	}

	// Значение не найдено в кэше, вызываем функцию
	result, err := fn()
	if err != nil {
		return "", err
	}

	// Сохраняем результат в кэше
	err = r.Set(ctx, key, result, expiration)
	if err != nil {
		log.Printf("Ошибка сохранения в кэше: %v", err)
	}

	// Преобразуем результат в строку
	strResult, ok := result.(string)
	if !ok {
		// Если результат не строка, сериализуем его в JSON
		jsonData, err := json.Marshal(result)
		if err != nil {
			return "", err
		}
		return string(jsonData), nil
	}

	return strResult, nil
}

// GetJSONOrSet получает объект из кэша или вызывает функцию для получения объекта и сохранения в кэше
func (r *RedisClient) GetJSONOrSet(ctx context.Context, key string, dest interface{}, expiration time.Duration, fn func() (interface{}, error)) error {
	// Пытаемся получить из кэша
	err := r.GetJSON(ctx, key, dest)
	if err == nil {
		// Значение найдено в кэше
		return nil
	}

	// Значение не найдено в кэше, вызываем функцию
	result, err := fn()
	if err != nil {
		return err
	}

	// Сохраняем результат в кэше
	return r.SetJSON(ctx, key, result, expiration)
}
