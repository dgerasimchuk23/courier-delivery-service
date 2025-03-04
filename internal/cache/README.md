# Пакет cache

Пакет `cache` предоставляет функциональность для работы с Redis в качестве кэша и хранилища временных данных.

## Основные компоненты

### RedisClient

`RedisClient` - основной клиент для работы с Redis, реализующий интерфейс `RedisClientInterface`.

```go
type RedisClient struct {
    client       *redis.Client
    cleanupDone  chan bool
    cleanupStats *RedisStats
}
```

### RedisClientInterface

Интерфейс, определяющий методы для работы с Redis:

```go
type RedisClientInterface interface {
    Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
    Get(ctx context.Context, key string) (string, error)
    Delete(ctx context.Context, key string) error
    Close() error
    SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error
    GetJSON(ctx context.Context, key string, dest interface{}) error
    MonitorStats(ctx context.Context) (*RedisStats, error)
    CleanupRateLimitKeys(ctx context.Context) (int64, error)
    CleanupBlacklistKeys(ctx context.Context) (int64, error)
    ScheduleRedisCleanup(interval time.Duration) chan bool
}
```

## Мониторинг и очистка Redis

### Мониторинг

Пакет предоставляет функциональность для мониторинга использования Redis:

```go
// RedisStats содержит статистику использования Redis
type RedisStats struct {
    TotalKeys       int64
    ExpiredKeys     int64
    EvictedKeys     int64
    UsedMemory      int64
    UsedMemoryPeak  int64
    RateLimitKeys   int64
    BlacklistKeys   int64
    CleanupDuration time.Duration
}
```

Метод `MonitorStats` позволяет получить текущую статистику использования Redis:

```go
stats, err := redisClient.MonitorStats(ctx)
if err != nil {
    log.Printf("Ошибка при получении статистики Redis: %v", err)
} else {
    log.Printf("Статистика Redis: total keys=%d, rate limit keys=%d, blacklist keys=%d", 
        stats.TotalKeys, stats.RateLimitKeys, stats.BlacklistKeys)
}
```

### Очистка

Пакет предоставляет методы для очистки устаревших ключей в Redis:

1. `CleanupRateLimitKeys` - удаляет устаревшие ключи rate limit
2. `CleanupBlacklistKeys` - удаляет устаревшие ключи из черного списка токенов

Для автоматической периодической очистки можно использовать метод `ScheduleRedisCleanup`:

```go
// Запускаем периодическую очистку Redis каждый час
cleanupDone := redisClient.ScheduleRedisCleanup(1 * time.Hour)

// При завершении работы приложения останавливаем очистку
defer func() {
    cleanupDone <- true
}()
```

## Использование в приложении

### Инициализация

```go
redisClient := cache.NewRedisClient(config)
if redisClient != nil {
    defer redisClient.Close()
    log.Println("Redis успешно инициализирован")
} else {
    log.Println("Не удалось инициализировать Redis, продолжаем без кэширования")
}
```

### Работа с кэшем

```go
// Сохранение значения в кэше
err := redisClient.Set(ctx, "key", "value", 10*time.Minute)

// Получение значения из кэша
value, err := redisClient.Get(ctx, "key")

// Удаление значения из кэша
err := redisClient.Delete(ctx, "key")
```

### Работа с JSON

```go
// Сохранение объекта в кэше
user := User{ID: 1, Name: "John"}
err := redisClient.SetJSON(ctx, "user:1", user, 10*time.Minute)

// Получение объекта из кэша
var user User
err := redisClient.GetJSON(ctx, "user:1", &user)
```

## Оптимизация работы с Redis

1. **Ограничение TTL для ключей** - все ключи должны иметь ограниченное время жизни, чтобы предотвратить переполнение Redis.
2. **Периодическая очистка** - используйте `ScheduleRedisCleanup` для периодической очистки устаревших ключей.
3. **Мониторинг использования** - регулярно проверяйте статистику использования Redis с помощью `MonitorStats`.
4. **Логирование** - все операции с Redis логируются для отладки и мониторинга.
5. **Обработка ошибок** - все методы возвращают ошибки, которые должны быть обработаны.

## Тестирование

Для тестирования пакета используются моки и юнит-тесты:

```go
func TestMonitorStats(t *testing.T) {
    // Создаем мок для redis.Client
    mockClient := new(MockRedisClient)
    
    // Настраиваем мок
    mockClient.On("Info", mock.Anything, []string{"stats", "memory"}).Return(
        "# Stats\r\nexpired_keys:100\r\nevicted_keys:50\r\n# Memory\r\nused_memory:1000\r\nused_memory_peak:2000\r\n",
        nil,
    )
    
    // Вызываем тестируемый метод
    ctx := context.Background()
    stats, err := redisClient.MonitorStats(ctx)
    
    // Проверяем результаты
    assert.NoError(t, err)
    assert.Equal(t, int64(100), stats.ExpiredKeys)
}
``` 