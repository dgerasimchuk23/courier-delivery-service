package cache

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

// RedisStats содержит статистику использования Redis
type RedisStats struct {
	// Общее количество ключей в Redis
	TotalKeys int64
	// Количество истекших ключей
	ExpiredKeys int64
	// Количество удаленных ключей из-за нехватки памяти
	EvictedKeys int64
	// Используемая память в байтах
	UsedMemory int64
	// Пиковая память, использованная Redis
	UsedMemoryPeak int64
	// Количество ключей, связанных с ограничением частоты запросов
	RateLimitKeys int64
	// Количество ключей в черном списке токенов
	BlacklistKeys int64
	// Время, затраченное на очистку устаревших ключей
	CleanupDuration time.Duration
}

// MonitorStats получает статистику использования Redis
func (r *RedisClient) MonitorStats(ctx context.Context) (*RedisStats, error) {
	if r == nil || r.client == nil {
		return nil, fmt.Errorf("Redis client is nil")
	}

	startTime := time.Now()
	stats := &RedisStats{}

	// Получаем общую информацию о Redis
	info, err := r.client.Info(ctx, "stats", "memory").Result()
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении информации о Redis: %w", err)
	}

	// Парсим информацию
	lines := strings.Split(info, "\r\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "expired_keys:") {
			if _, err := fmt.Sscanf(line, "expired_keys:%d", &stats.ExpiredKeys); err != nil {
				log.Printf("Ошибка при парсинге expired_keys: %v", err)
			}
		} else if strings.HasPrefix(line, "evicted_keys:") {
			if _, err := fmt.Sscanf(line, "evicted_keys:%d", &stats.EvictedKeys); err != nil {
				log.Printf("Ошибка при парсинге evicted_keys: %v", err)
			}
		} else if strings.HasPrefix(line, "used_memory:") {
			if _, err := fmt.Sscanf(line, "used_memory:%d", &stats.UsedMemory); err != nil {
				log.Printf("Ошибка при парсинге used_memory: %v", err)
			}
		} else if strings.HasPrefix(line, "used_memory_peak:") {
			if _, err := fmt.Sscanf(line, "used_memory_peak:%d", &stats.UsedMemoryPeak); err != nil {
				log.Printf("Ошибка при парсинге used_memory_peak: %v", err)
			}
		}
	}

	// Получаем общее количество ключей
	dbSize, err := r.client.DBSize(ctx).Result()
	if err != nil {
		log.Printf("Ошибка при получении размера базы данных Redis: %v", err)
	} else {
		stats.TotalKeys = dbSize
	}

	// Получаем количество ключей для rate limiting
	rateLimitKeys, err := r.client.Keys(ctx, "rate_limit:*").Result()
	if err != nil {
		log.Printf("Ошибка при получении ключей rate limit: %v", err)
	} else {
		stats.RateLimitKeys = int64(len(rateLimitKeys))
	}

	// Получаем количество ключей для черного списка токенов
	blacklistKeys, err := r.client.Keys(ctx, "blacklist:*").Result()
	if err != nil {
		log.Printf("Ошибка при получении ключей blacklist: %v", err)
	} else {
		stats.BlacklistKeys = int64(len(blacklistKeys))
	}

	stats.CleanupDuration = time.Since(startTime)
	return stats, nil
}

// CleanupRateLimitKeys удаляет устаревшие ключи rate limit
func (r *RedisClient) CleanupRateLimitKeys(ctx context.Context) (int64, error) {
	if r == nil || r.client == nil {
		return 0, fmt.Errorf("Redis client is nil")
	}

	// Получаем все ключи rate limit
	keys, err := r.client.Keys(ctx, "rate_limit:*").Result()
	if err != nil {
		return 0, fmt.Errorf("ошибка при получении ключей rate limit: %w", err)
	}

	// Проверяем TTL для каждого ключа и удаляем те, у которых TTL < 0 (истекшие)
	var deleted int64
	for _, key := range keys {
		ttl, err := r.client.TTL(ctx, key).Result()
		if err != nil {
			log.Printf("Ошибка при получении TTL для ключа %s: %v", key, err)
			continue
		}

		// Если TTL < 0, ключ уже истек или не имеет TTL
		if ttl < 0 {
			_, err := r.client.Del(ctx, key).Result()
			if err != nil {
				log.Printf("Ошибка при удалении ключа %s: %v", key, err)
				continue
			}
			deleted++
			log.Printf("Удален устаревший ключ rate limit: %s", key)
		}
	}

	return deleted, nil
}

// CleanupBlacklistKeys удаляет устаревшие ключи из черного списка токенов
func (r *RedisClient) CleanupBlacklistKeys(ctx context.Context) (int64, error) {
	if r == nil || r.client == nil {
		return 0, fmt.Errorf("Redis client is nil")
	}

	// Получаем все ключи blacklist
	keys, err := r.client.Keys(ctx, "blacklist:*").Result()
	if err != nil {
		return 0, fmt.Errorf("ошибка при получении ключей blacklist: %w", err)
	}

	// Проверяем TTL для каждого ключа и удаляем те, у которых TTL < 0 (истекшие)
	var deleted int64
	for _, key := range keys {
		ttl, err := r.client.TTL(ctx, key).Result()
		if err != nil {
			log.Printf("Ошибка при получении TTL для ключа %s: %v", key, err)
			continue
		}

		// Если TTL < 0, ключ уже истек или не имеет TTL
		if ttl < 0 {
			_, err := r.client.Del(ctx, key).Result()
			if err != nil {
				log.Printf("Ошибка при удалении ключа %s: %v", key, err)
				continue
			}
			deleted++
			log.Printf("Удален устаревший ключ blacklist: %s", key)
		}
	}

	return deleted, nil
}

// ScheduleRedisCleanup запускает периодическую очистку Redis
func (r *RedisClient) ScheduleRedisCleanup(interval time.Duration) chan bool {
	done := make(chan bool)
	ticker := time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-ticker.C:
				ctx := context.Background()

				// Получаем статистику перед очисткой
				statsBefore, err := r.MonitorStats(ctx)
				if err != nil {
					log.Printf("Ошибка при получении статистики Redis: %v", err)
					continue
				}

				// Очищаем устаревшие ключи rate limit
				deletedRateLimit, err := r.CleanupRateLimitKeys(ctx)
				if err != nil {
					log.Printf("Ошибка при очистке ключей rate limit: %v", err)
				}

				// Очищаем устаревшие ключи blacklist
				deletedBlacklist, err := r.CleanupBlacklistKeys(ctx)
				if err != nil {
					log.Printf("Ошибка при очистке ключей blacklist: %v", err)
				}

				// Получаем статистику после очистки
				statsAfter, err := r.MonitorStats(ctx)
				if err != nil {
					log.Printf("Ошибка при получении статистики Redis: %v", err)
					continue
				}

				// Логируем результаты очистки
				log.Printf("Очистка Redis завершена: удалено %d ключей rate limit, %d ключей blacklist",
					deletedRateLimit, deletedBlacklist)
				log.Printf("Статистика Redis до очистки: total keys=%d, rate limit keys=%d, blacklist keys=%d",
					statsBefore.TotalKeys, statsBefore.RateLimitKeys, statsBefore.BlacklistKeys)
				log.Printf("Статистика Redis после очистки: total keys=%d, rate limit keys=%d, blacklist keys=%d",
					statsAfter.TotalKeys, statsAfter.RateLimitKeys, statsAfter.BlacklistKeys)

			case <-done:
				ticker.Stop()
				return
			}
		}
	}()

	return done
}
