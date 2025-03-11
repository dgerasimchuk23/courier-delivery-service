package middleware

import (
	"context"
	"delivery/internal/cache"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// RateLimitConfig содержит конфигурацию для ограничения частоты запросов
type RateLimitConfig struct {
	// Лимиты для аутентифицированных пользователей
	AuthenticatedLimits map[string]int `json:"authenticated_limits"`
	// Лимит для неаутентифицированных пользователей
	UnauthenticatedLimit int `json:"unauthenticated_limit"`
	// Время блокировки при превышении лимита (в минутах)
	BlockDuration int `json:"block_duration"`
}

// DefaultRateLimitConfig возвращает конфигурацию по умолчанию
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		AuthenticatedLimits: map[string]int{
			"client":  30,  // Клиенты: 30 запросов в минуту
			"courier": 137, // Курьеры: 137 запросов в минуту
		},
		UnauthenticatedLimit: 15, // Неаутентифицированные: 15 запросов в минуту
		BlockDuration:        1,  // Блокировка на 1 минуту при превышении лимита
	}
}

// RateLimiter представляет middleware для ограничения частоты запросов
type RateLimiter struct {
	redisClient cache.RedisClientInterface
	config      RateLimitConfig
}

// NewRateLimiter создает новый экземпляр RateLimiter
func NewRateLimiter(redisClient cache.RedisClientInterface, config RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		redisClient: redisClient,
		config:      config,
	}
}

// LoadConfigFromRedis загружает конфигурацию из Redis
func (rl *RateLimiter) LoadConfigFromRedis(ctx context.Context) error {
	configJSON, err := rl.redisClient.Get(ctx, "rate_limit_config")
	if err != nil {
		// Если конфигурация не найдена, используем значения по умолчанию
		log.Println("Конфигурация Rate Limiting не найдена в Redis, используем значения по умолчанию")
		return nil
	}

	var config RateLimitConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return fmt.Errorf("ошибка при десериализации конфигурации Rate Limiting: %w", err)
	}

	rl.config = config
	log.Println("Конфигурация Rate Limiting загружена из Redis")
	return nil
}

// SaveConfigToRedis сохраняет конфигурацию в Redis
func (rl *RateLimiter) SaveConfigToRedis(ctx context.Context) error {
	configJSON, err := json.Marshal(rl.config)
	if err != nil {
		return fmt.Errorf("ошибка при сериализации конфигурации Rate Limiting: %w", err)
	}

	if err := rl.redisClient.Set(ctx, "rate_limit_config", string(configJSON), 0); err != nil {
		return fmt.Errorf("ошибка при сохранении конфигурации Rate Limiting в Redis: %w", err)
	}

	log.Println("Конфигурация Rate Limiting сохранена в Redis")
	return nil
}

// UpdateConfig обновляет конфигурацию и сохраняет ее в Redis
func (rl *RateLimiter) UpdateConfig(config RateLimitConfig) error {
	rl.config = config
	ctx := context.Background()
	return rl.SaveConfigToRedis(ctx)
}

// getClientIP получает IP-адрес клиента из запроса
func getClientIP(r *http.Request) string {
	// Проверяем заголовки X-Forwarded-For и X-Real-IP
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return xrip
	}

	// Получаем IP из RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// getUserID получает ID пользователя из контекста запроса
func getUserID(r *http.Request) (int, bool) {
	userID, ok := r.Context().Value(UserIDKey).(int)
	return userID, ok
}

// getUserRole получает роль пользователя из контекста запроса
func getUserRole(r *http.Request) (string, bool) {
	role, ok := r.Context().Value(UserRoleKey).(string)
	return role, ok
}

// Middleware возвращает middleware для ограничения частоты запросов
func (rl *RateLimiter) Middleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Если Redis недоступен, пропускаем запрос
			if rl.redisClient == nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := r.Context()

			// Получаем IP-адрес клиента
			clientIP := getClientIP(r)

			// Проверяем, не заблокирован ли IP
			blockKey := fmt.Sprintf("rate_limit:block:%s", clientIP)
			blockValue, err := rl.redisClient.Get(ctx, blockKey)
			if err == nil {
				// IP заблокирован
				w.Header().Set("Retry-After", strconv.Itoa(rl.config.BlockDuration*60))
				http.Error(w, "Слишком много запросов. Пожалуйста, повторите попытку позже.", http.StatusTooManyRequests)

				// Логируем блокировку с дополнительной информацией
				log.Printf("[RATE_LIMIT_BLOCKED] IP %s заблокирован, значение: %s, URL: %s, метод: %s",
					clientIP, blockValue, r.URL.Path, r.Method)
				return
			}

			// Определяем лимит в зависимости от аутентификации
			var limit int
			var keyPrefix string
			var userInfo string

			userID, authenticated := getUserID(r)
			if authenticated {
				// Пользователь аутентифицирован
				role, hasRole := getUserRole(r)
				if hasRole {
					// Используем лимит в зависимости от роли
					if roleLimit, ok := rl.config.AuthenticatedLimits[role]; ok {
						limit = roleLimit
						keyPrefix = fmt.Sprintf("rate_limit:auth:%s:%d", role, userID)
						userInfo = fmt.Sprintf("пользователь %d с ролью %s", userID, role)
					} else {
						// Если роль не найдена, используем лимит для клиентов
						limit = rl.config.AuthenticatedLimits["client"]
						keyPrefix = fmt.Sprintf("rate_limit:auth:client:%d", userID)
						userInfo = fmt.Sprintf("пользователь %d с неизвестной ролью (используем client)", userID)
					}
				} else {
					// Если роль не определена, используем лимит для клиентов
					limit = rl.config.AuthenticatedLimits["client"]
					keyPrefix = fmt.Sprintf("rate_limit:auth:client:%d", userID)
					userInfo = fmt.Sprintf("пользователь %d без роли", userID)
				}
			} else {
				// Пользователь не аутентифицирован
				limit = rl.config.UnauthenticatedLimit
				keyPrefix = fmt.Sprintf("rate_limit:unauth:%s", clientIP)
				userInfo = fmt.Sprintf("неаутентифицированный IP %s", clientIP)
			}

			// Получаем текущее количество запросов
			countKey := fmt.Sprintf("%s:count", keyPrefix)
			countStr, err := rl.redisClient.Get(ctx, countKey)
			var count int
			if err == nil {
				count, _ = strconv.Atoi(countStr)
			}

			// Увеличиваем счетчик
			count++

			// Если это первый запрос, устанавливаем TTL
			var setErr error
			if count == 1 {
				setErr = rl.redisClient.Set(ctx, countKey, strconv.Itoa(count), time.Minute)
				if setErr == nil {
					log.Printf("[RATE_LIMIT_NEW] Новый счетчик для %s: %d/%d, URL: %s, метод: %s",
						userInfo, count, limit, r.URL.Path, r.Method)
				}
			} else {
				// Иначе просто обновляем значение (TTL сохраняется)
				setErr = rl.redisClient.Set(ctx, countKey, strconv.Itoa(count), 0)

				// Логируем каждый 10-й запрос или если счетчик приближается к лимиту
				if setErr == nil && (count%10 == 0 || count > limit*80/100) {
					log.Printf("[RATE_LIMIT_UPDATE] Обновлен счетчик для %s: %d/%d (%.1f%%), URL: %s, метод: %s",
						userInfo, count, limit, float64(count)/float64(limit)*100, r.URL.Path, r.Method)
				}
			}

			if setErr != nil {
				log.Printf("[RATE_LIMIT_ERROR] Ошибка при обновлении счетчика запросов: %v", setErr)
				// В случае ошибки Redis пропускаем запрос
				next.ServeHTTP(w, r)
				return
			}

			// Проверяем, не превышен ли лимит
			if count > limit {
				// Если пользователь не аутентифицирован, блокируем IP на указанное время
				if !authenticated {
					blockDuration := time.Duration(rl.config.BlockDuration) * time.Minute
					blockValue := fmt.Sprintf("blocked:count=%d:limit=%d:time=%s",
						count, limit, time.Now().Format(time.RFC3339))
					err = rl.redisClient.Set(ctx, blockKey, blockValue, blockDuration)
					if err != nil {
						log.Printf("[RATE_LIMIT_ERROR] Ошибка при блокировке IP: %v", err)
					} else {
						log.Printf("[RATE_LIMIT_BLOCK] IP %s заблокирован на %v: %d/%d запросов, URL: %s, метод: %s",
							clientIP, blockDuration, count, limit, r.URL.Path, r.Method)
					}
				} else {
					log.Printf("[RATE_LIMIT_EXCEEDED] %s превысил лимит: %d/%d запросов, URL: %s, метод: %s",
						userInfo, count, limit, r.URL.Path, r.Method)
				}

				// Возвращаем ошибку 429 Too Many Requests
				w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("Retry-After", strconv.Itoa(60)) // Retry after 60 seconds
				http.Error(w, "Слишком много запросов. Пожалуйста, повторите попытку позже.", http.StatusTooManyRequests)
				return
			}

			// Устанавливаем заголовки с информацией о лимите
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(limit-count))

			// Передаем запрос следующему обработчику
			next.ServeHTTP(w, r)
		})
	}
}
