package cache

import (
	"context"
	"log"
	"time"
)

// LogInitialStats выводит начальную статистику Redis
func (c *RedisClient) LogInitialStats() {
	if c == nil {
		log.Println("Redis client is nil, skipping stats logging")
		return
	}

	log.Println("Redis успешно инициализирован")

	// Выводим начальную статистику Redis
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stats, err := c.MonitorStats(ctx)
	if err != nil {
		log.Printf("Ошибка при получении статистики Redis: %v", err)
	} else {
		log.Printf("Статистика Redis при запуске: total keys=%d, rate limit keys=%d, blacklist keys=%d",
			stats.TotalKeys, stats.RateLimitKeys, stats.BlacklistKeys)
		log.Printf("Память Redis: used=%d bytes, peak=%d bytes",
			stats.UsedMemory, stats.UsedMemoryPeak)
	}
}

// LogFinalStats выводит финальную статистику Redis
func (c *RedisClient) LogFinalStats() {
	if c == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stats, err := c.MonitorStats(ctx)
	if err != nil {
		log.Printf("Ошибка при получении финальной статистики Redis: %v", err)
	} else {
		log.Printf("Финальная статистика Redis: total keys=%d, rate limit keys=%d, blacklist keys=%d",
			stats.TotalKeys, stats.RateLimitKeys, stats.BlacklistKeys)
		log.Printf("Память Redis: used=%d bytes, peak=%d bytes",
			stats.UsedMemory, stats.UsedMemoryPeak)
	}
}
