# Middleware Package

Пакет `middleware` содержит компоненты промежуточного программного обеспечения (middleware) для HTTP-запросов в приложении доставки.

## Компоненты

### AuthMiddleware

`AuthMiddleware` отвечает за аутентификацию пользователей. Он проверяет JWT-токены в заголовке `Authorization` и добавляет информацию о пользователе в контекст запроса.

#### Основные функции:

- `NewAuthMiddleware(authService auth.AuthServiceInterface) *AuthMiddleware` - создает новый экземпляр AuthMiddleware
- `Middleware() mux.MiddlewareFunc` - возвращает middleware для аутентификации

### RateLimiter

`RateLimiter` ограничивает частоту запросов от клиентов. Он использует Redis для хранения счетчиков запросов и блокировки IP-адресов, превышающих лимиты.

#### Основные функции:

- `NewRateLimiter(redisClient cache.RedisClientInterface, config RateLimitConfig) *RateLimiter` - создает новый экземпляр RateLimiter
- `Middleware() mux.MiddlewareFunc` - возвращает middleware для ограничения частоты запросов
- `LoadConfigFromRedis(ctx context.Context) error` - загружает конфигурацию из Redis
- `SaveConfigToRedis(ctx context.Context) error` - сохраняет конфигурацию в Redis
- `UpdateConfig(config RateLimitConfig) error` - обновляет конфигурацию и сохраняет ее в Redis

#### Конфигурация:

```go
type RateLimitConfig struct {
    // Лимиты для аутентифицированных пользователей
    AuthenticatedLimits map[string]int `json:"authenticated_limits"`
    // Лимит для неаутентифицированных пользователей
    UnauthenticatedLimit int `json:"unauthenticated_limit"`
    // Время блокировки при превышении лимита (в минутах)
    BlockDuration int `json:"block_duration"`
}
```

## Использование

```go
// Инициализация middleware
authMiddleware := middleware.NewAuthMiddleware(authService)
rateLimiter := middleware.NewRateLimiter(redisClient, middleware.DefaultRateLimitConfig())

// Применение middleware к маршрутизатору
router := mux.NewRouter()
router.Use(authMiddleware.Middleware())
router.Use(rateLimiter.Middleware())
```

## Тестирование

Для тестирования middleware используются моки для зависимостей:

- `MockAuthService` - мок для `auth.AuthServiceInterface`
- `MockRedisClient` - мок для `cache.RedisClientInterface`

Примеры тестов можно найти в файлах `auth_middleware_test.go` и `rate_limiter_test.go`.

## TODO

- Добавить получение роли пользователя из базы данных в `AuthMiddleware`
- Заменить трекинг заказа и геотрекинг на WebSockets для более эффективной работы 