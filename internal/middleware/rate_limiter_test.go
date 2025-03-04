package middleware

import (
	"context"
	"delivery/internal/cache"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRedisClient - мок для RedisClientInterface
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockRedisClient) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *MockRedisClient) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *MockRedisClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRedisClient) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockRedisClient) GetJSON(ctx context.Context, key string, dest interface{}) error {
	args := m.Called(ctx, key, dest)
	return args.Error(0)
}

// MonitorStats возвращает статистику использования Redis
func (m *MockRedisClient) MonitorStats(ctx context.Context) (*cache.RedisStats, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*cache.RedisStats), args.Error(1)
}

// CleanupRateLimitKeys удаляет устаревшие ключи rate limit
func (m *MockRedisClient) CleanupRateLimitKeys(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// CleanupBlacklistKeys удаляет устаревшие ключи из черного списка токенов
func (m *MockRedisClient) CleanupBlacklistKeys(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// ScheduleRedisCleanup запускает периодическую очистку Redis
func (m *MockRedisClient) ScheduleRedisCleanup(interval time.Duration) chan bool {
	args := m.Called(interval)
	return args.Get(0).(chan bool)
}

func TestRateLimiter_Middleware(t *testing.T) {
	// Создаем мок для RedisClient
	mockRedis := new(MockRedisClient)

	// Создаем конфигурацию для тестов
	config := RateLimitConfig{
		AuthenticatedLimits: map[string]int{
			"client":  30,
			"courier": 137,
		},
		UnauthenticatedLimit: 15,
		BlockDuration:        1,
	}

	// Создаем RateLimiter с мок-клиентом Redis
	rateLimiter := NewRateLimiter(mockRedis, config)

	// Создаем тестовый обработчик
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Создаем middleware
	middleware := rateLimiter.Middleware()

	// Создаем тестовый сервер
	router := mux.NewRouter()
	router.Use(middleware)
	router.HandleFunc("/test", testHandler).Methods("GET")
	server := httptest.NewServer(router)
	defer server.Close()

	t.Run("Неаутентифицированный пользователь не превышает лимит", func(t *testing.T) {
		// Настраиваем мок для проверки блокировки IP
		mockRedis.On("Get", mock.Anything, "rate_limit:block:127.0.0.1").Return("", cache.ErrKeyNotFound).Once()

		// Настраиваем мок для получения счетчика запросов
		mockRedis.On("Get", mock.Anything, "rate_limit:unauth:127.0.0.1:count").Return("", cache.ErrKeyNotFound).Once()

		// Настраиваем мок для установки счетчика запросов
		mockRedis.On("Set", mock.Anything, "rate_limit:unauth:127.0.0.1:count", "1", time.Minute).Return(nil).Once()

		// Создаем тестовый запрос
		req, err := http.NewRequest("GET", "/test", nil)
		assert.NoError(t, err)
		req.RemoteAddr = "127.0.0.1:1234"

		// Создаем ResponseRecorder для записи ответа
		rr := httptest.NewRecorder()

		// Выполняем запрос
		router.ServeHTTP(rr, req)

		// Проверяем статус ответа
		assert.Equal(t, http.StatusOK, rr.Code)

		// Проверяем заголовки
		assert.Equal(t, "15", rr.Header().Get("X-RateLimit-Limit"))
		assert.Equal(t, "14", rr.Header().Get("X-RateLimit-Remaining"))

		// Проверяем, что все ожидаемые вызовы были выполнены
		mockRedis.AssertExpectations(t)
	})

	t.Run("Неаутентифицированный пользователь превышает лимит", func(t *testing.T) {
		// Сбрасываем мок
		mockRedis = new(MockRedisClient)
		rateLimiter = NewRateLimiter(mockRedis, config)

		// Обновляем middleware и router
		middleware = rateLimiter.Middleware()
		router = mux.NewRouter()
		router.Use(middleware)
		router.HandleFunc("/test", testHandler).Methods("GET")

		// Настраиваем мок для проверки блокировки IP
		mockRedis.On("Get", mock.Anything, "rate_limit:block:127.0.0.1").Return("", cache.ErrKeyNotFound).Once()

		// Настраиваем мок для получения счетчика запросов
		mockRedis.On("Get", mock.Anything, "rate_limit:unauth:127.0.0.1:count").Return("15", nil).Once()

		// Настраиваем мок для установки счетчика запросов
		mockRedis.On("Set", mock.Anything, "rate_limit:unauth:127.0.0.1:count", "16", time.Duration(0)).Return(nil).Once()

		// Настраиваем мок для блокировки IP
		mockRedis.On("Set", mock.Anything, "rate_limit:block:127.0.0.1", mock.Anything, time.Minute).Return(nil).Once()

		// Создаем тестовый запрос
		req, err := http.NewRequest("GET", "/test", nil)
		assert.NoError(t, err)
		req.RemoteAddr = "127.0.0.1:1234"

		// Создаем ResponseRecorder для записи ответа
		rr := httptest.NewRecorder()

		// Выполняем запрос
		router.ServeHTTP(rr, req)

		// Проверяем статус ответа
		assert.Equal(t, http.StatusTooManyRequests, rr.Code)

		// Проверяем заголовки
		assert.Equal(t, "15", rr.Header().Get("X-RateLimit-Limit"))
		assert.Equal(t, "0", rr.Header().Get("X-RateLimit-Remaining"))
		assert.Equal(t, "60", rr.Header().Get("Retry-After"))

		// Проверяем, что все ожидаемые вызовы были выполнены
		mockRedis.AssertExpectations(t)
	})

	t.Run("Аутентифицированный пользователь (клиент) не превышает лимит", func(t *testing.T) {
		// Сбрасываем мок
		mockRedis = new(MockRedisClient)
		rateLimiter = NewRateLimiter(mockRedis, config)

		// Обновляем middleware и router
		middleware = rateLimiter.Middleware()
		router = mux.NewRouter()
		router.Use(middleware)
		router.HandleFunc("/test", testHandler).Methods("GET")

		// Настраиваем мок для проверки блокировки IP
		mockRedis.On("Get", mock.Anything, "rate_limit:block:127.0.0.1").Return("", cache.ErrKeyNotFound).Once()

		// Настраиваем мок для получения счетчика запросов
		mockRedis.On("Get", mock.Anything, "rate_limit:auth:client:123:count").Return("", cache.ErrKeyNotFound).Once()

		// Настраиваем мок для установки счетчика запросов
		mockRedis.On("Set", mock.Anything, "rate_limit:auth:client:123:count", "1", time.Minute).Return(nil).Once()

		// Создаем тестовый запрос с аутентификацией
		req, err := http.NewRequest("GET", "/test", nil)
		assert.NoError(t, err)
		req.RemoteAddr = "127.0.0.1:1234"

		// Добавляем информацию о пользователе в контекст
		ctx := context.WithValue(req.Context(), "user_id", 123)
		ctx = context.WithValue(ctx, "user_role", "client")
		req = req.WithContext(ctx)

		// Создаем ResponseRecorder для записи ответа
		rr := httptest.NewRecorder()

		// Выполняем запрос
		router.ServeHTTP(rr, req)

		// Проверяем статус ответа
		assert.Equal(t, http.StatusOK, rr.Code)

		// Проверяем заголовки
		assert.Equal(t, "30", rr.Header().Get("X-RateLimit-Limit"))
		assert.Equal(t, "29", rr.Header().Get("X-RateLimit-Remaining"))

		// Проверяем, что все ожидаемые вызовы были выполнены
		mockRedis.AssertExpectations(t)
	})

	t.Run("Аутентифицированный пользователь (курьер) не превышает лимит", func(t *testing.T) {
		// Сбрасываем мок
		mockRedis = new(MockRedisClient)
		rateLimiter = NewRateLimiter(mockRedis, config)

		// Обновляем middleware и router
		middleware = rateLimiter.Middleware()
		router = mux.NewRouter()
		router.Use(middleware)
		router.HandleFunc("/test", testHandler).Methods("GET")

		// Настраиваем мок для проверки блокировки IP
		mockRedis.On("Get", mock.Anything, "rate_limit:block:127.0.0.1").Return("", cache.ErrKeyNotFound).Once()

		// Настраиваем мок для получения счетчика запросов
		mockRedis.On("Get", mock.Anything, "rate_limit:auth:courier:456:count").Return("", cache.ErrKeyNotFound).Once()

		// Настраиваем мок для установки счетчика запросов
		mockRedis.On("Set", mock.Anything, "rate_limit:auth:courier:456:count", "1", time.Minute).Return(nil).Once()

		// Создаем тестовый запрос с аутентификацией
		req, err := http.NewRequest("GET", "/test", nil)
		assert.NoError(t, err)
		req.RemoteAddr = "127.0.0.1:1234"

		// Добавляем информацию о пользователе в контекст
		ctx := context.WithValue(req.Context(), "user_id", 456)
		ctx = context.WithValue(ctx, "user_role", "courier")
		req = req.WithContext(ctx)

		// Создаем ResponseRecorder для записи ответа
		rr := httptest.NewRecorder()

		// Выполняем запрос
		router.ServeHTTP(rr, req)

		// Проверяем статус ответа
		assert.Equal(t, http.StatusOK, rr.Code)

		// Проверяем заголовки
		assert.Equal(t, "137", rr.Header().Get("X-RateLimit-Limit"))
		assert.Equal(t, "136", rr.Header().Get("X-RateLimit-Remaining"))

		// Проверяем, что все ожидаемые вызовы были выполнены
		mockRedis.AssertExpectations(t)
	})

	t.Run("IP заблокирован", func(t *testing.T) {
		// Сбрасываем мок
		mockRedis = new(MockRedisClient)
		rateLimiter = NewRateLimiter(mockRedis, config)

		// Обновляем middleware и router
		middleware = rateLimiter.Middleware()
		router = mux.NewRouter()
		router.Use(middleware)
		router.HandleFunc("/test", testHandler).Methods("GET")

		// Настраиваем мок для проверки блокировки IP
		mockRedis.On("Get", mock.Anything, "rate_limit:block:127.0.0.1").Return("blocked", nil).Once()

		// Создаем тестовый запрос
		req, err := http.NewRequest("GET", "/test", nil)
		assert.NoError(t, err)
		req.RemoteAddr = "127.0.0.1:1234"

		// Создаем ResponseRecorder для записи ответа
		rr := httptest.NewRecorder()

		// Выполняем запрос
		router.ServeHTTP(rr, req)

		// Проверяем статус ответа
		assert.Equal(t, http.StatusTooManyRequests, rr.Code)

		// Проверяем заголовки
		assert.Equal(t, "60", rr.Header().Get("Retry-After"))

		// Проверяем, что все ожидаемые вызовы были выполнены
		mockRedis.AssertExpectations(t)
	})
}

func TestRateLimiter_LoadConfigFromRedis(t *testing.T) {
	// Создаем мок для RedisClient
	mockRedis := new(MockRedisClient)

	// Создаем конфигурацию для тестов
	config := DefaultRateLimitConfig()

	// Создаем RateLimiter с мок-клиентом Redis
	rateLimiter := NewRateLimiter(mockRedis, config)

	t.Run("Конфигурация успешно загружена из Redis", func(t *testing.T) {
		// Настраиваем мок для получения конфигурации из Redis
		configJSON := `{"authenticated_limits":{"client":50,"courier":200},"unauthenticated_limit":20,"block_duration":2}`
		mockRedis.On("Get", mock.Anything, "rate_limit_config").Return(configJSON, nil).Once()

		// Загружаем конфигурацию из Redis
		err := rateLimiter.LoadConfigFromRedis(context.Background())
		assert.NoError(t, err)

		// Проверяем, что конфигурация была обновлена
		assert.Equal(t, 50, rateLimiter.config.AuthenticatedLimits["client"])
		assert.Equal(t, 200, rateLimiter.config.AuthenticatedLimits["courier"])
		assert.Equal(t, 20, rateLimiter.config.UnauthenticatedLimit)
		assert.Equal(t, 2, rateLimiter.config.BlockDuration)

		// Проверяем, что все ожидаемые вызовы были выполнены
		mockRedis.AssertExpectations(t)
	})

	t.Run("Конфигурация не найдена в Redis", func(t *testing.T) {
		// Сбрасываем мок
		mockRedis = new(MockRedisClient)
		rateLimiter = NewRateLimiter(mockRedis, config)

		// Настраиваем мок для получения конфигурации из Redis
		mockRedis.On("Get", mock.Anything, "rate_limit_config").Return("", cache.ErrKeyNotFound).Once()

		// Загружаем конфигурацию из Redis
		err := rateLimiter.LoadConfigFromRedis(context.Background())
		assert.NoError(t, err)

		// Проверяем, что конфигурация не изменилась
		assert.Equal(t, 30, rateLimiter.config.AuthenticatedLimits["client"])
		assert.Equal(t, 137, rateLimiter.config.AuthenticatedLimits["courier"])
		assert.Equal(t, 15, rateLimiter.config.UnauthenticatedLimit)
		assert.Equal(t, 1, rateLimiter.config.BlockDuration)

		// Проверяем, что все ожидаемые вызовы были выполнены
		mockRedis.AssertExpectations(t)
	})

	t.Run("Ошибка при десериализации конфигурации", func(t *testing.T) {
		// Сбрасываем мок
		mockRedis = new(MockRedisClient)
		rateLimiter = NewRateLimiter(mockRedis, config)

		// Настраиваем мок для получения конфигурации из Redis
		mockRedis.On("Get", mock.Anything, "rate_limit_config").Return("invalid json", nil).Once()

		// Загружаем конфигурацию из Redis
		err := rateLimiter.LoadConfigFromRedis(context.Background())
		assert.Error(t, err)

		// Проверяем, что все ожидаемые вызовы были выполнены
		mockRedis.AssertExpectations(t)
	})
}

func TestRateLimiter_SaveConfigToRedis(t *testing.T) {
	// Создаем мок для RedisClient
	mockRedis := new(MockRedisClient)

	// Создаем конфигурацию для тестов
	config := RateLimitConfig{
		AuthenticatedLimits: map[string]int{
			"client":  50,
			"courier": 200,
		},
		UnauthenticatedLimit: 20,
		BlockDuration:        2,
	}

	// Создаем RateLimiter с мок-клиентом Redis
	rateLimiter := NewRateLimiter(mockRedis, config)

	t.Run("Конфигурация успешно сохранена в Redis", func(t *testing.T) {
		// Настраиваем мок для сохранения конфигурации в Redis
		mockRedis.On("Set", mock.Anything, "rate_limit_config", mock.Anything, time.Duration(0)).Return(nil).Once()

		// Сохраняем конфигурацию в Redis
		err := rateLimiter.SaveConfigToRedis(context.Background())
		assert.NoError(t, err)

		// Проверяем, что все ожидаемые вызовы были выполнены
		mockRedis.AssertExpectations(t)
	})

	t.Run("Ошибка при сохранении конфигурации в Redis", func(t *testing.T) {
		// Сбрасываем мок
		mockRedis = new(MockRedisClient)
		rateLimiter = NewRateLimiter(mockRedis, config)

		// Настраиваем мок для сохранения конфигурации в Redis
		mockRedis.On("Set", mock.Anything, "rate_limit_config", mock.Anything, time.Duration(0)).Return(assert.AnError).Once()

		// Сохраняем конфигурацию в Redis
		err := rateLimiter.SaveConfigToRedis(context.Background())
		assert.Error(t, err)

		// Проверяем, что все ожидаемые вызовы были выполнены
		mockRedis.AssertExpectations(t)
	})
}

func TestRateLimiter_UpdateConfig(t *testing.T) {
	// Создаем мок для RedisClient
	mockRedis := new(MockRedisClient)

	// Создаем конфигурацию для тестов
	config := DefaultRateLimitConfig()

	// Создаем RateLimiter с мок-клиентом Redis
	rateLimiter := NewRateLimiter(mockRedis, config)

	t.Run("Конфигурация успешно обновлена", func(t *testing.T) {
		// Создаем новую конфигурацию
		newConfig := RateLimitConfig{
			AuthenticatedLimits: map[string]int{
				"client":  50,
				"courier": 200,
			},
			UnauthenticatedLimit: 20,
			BlockDuration:        2,
		}

		// Настраиваем мок для сохранения конфигурации в Redis
		mockRedis.On("Set", mock.Anything, "rate_limit_config", mock.Anything, time.Duration(0)).Return(nil).Once()

		// Обновляем конфигурацию
		err := rateLimiter.UpdateConfig(newConfig)
		assert.NoError(t, err)

		// Проверяем, что конфигурация была обновлена
		assert.Equal(t, 50, rateLimiter.config.AuthenticatedLimits["client"])
		assert.Equal(t, 200, rateLimiter.config.AuthenticatedLimits["courier"])
		assert.Equal(t, 20, rateLimiter.config.UnauthenticatedLimit)
		assert.Equal(t, 2, rateLimiter.config.BlockDuration)

		// Проверяем, что все ожидаемые вызовы были выполнены
		mockRedis.AssertExpectations(t)
	})
}
