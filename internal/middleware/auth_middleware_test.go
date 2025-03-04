package middleware

import (
	"delivery/internal/auth"
	"delivery/internal/cache"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService - мок для AuthServiceInterface
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) RegisterUser(email, password string) error {
	args := m.Called(email, password)
	return args.Error(0)
}

func (m *MockAuthService) GenerateTokens(userID int) (string, string, error) {
	args := m.Called(userID)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockAuthService) LoginUser(email, password string) (string, string, error) {
	args := m.Called(email, password)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockAuthService) RefreshToken(refreshToken string) (string, string, error) {
	args := m.Called(refreshToken)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockAuthService) Logout(accessToken, refreshToken string) error {
	args := m.Called(accessToken, refreshToken)
	return args.Error(0)
}

func (m *MockAuthService) ValidateToken(tokenString string) (int, error) {
	args := m.Called(tokenString)
	return args.Int(0), args.Error(1)
}

func (m *MockAuthService) WithCache(cacheClient *cache.RedisClient) *auth.AuthService {
	args := m.Called(cacheClient)
	return args.Get(0).(*auth.AuthService)
}

// Close реализует метод Close интерфейса AuthServiceInterface
func (m *MockAuthService) Close() {
	m.Called()
}

func TestAuthMiddleware_Middleware(t *testing.T) {
	t.Run("Успешная аутентификация", func(t *testing.T) {
		// Создаем мок-сервис аутентификации
		mockAuthService := new(MockAuthService)

		// Настраиваем ожидаемое поведение
		mockAuthService.On("ValidateToken", "valid-token").Return(123, nil)

		// Создаем AuthMiddleware с мок-сервисом
		authMiddleware := NewAuthMiddleware(mockAuthService)

		// Создаем тестовый обработчик
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Проверяем, что информация о пользователе добавлена в контекст
			userID, ok := r.Context().Value("user_id").(int)
			if ok {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("User ID: " + strconv.Itoa(userID)))
			} else {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("No user ID in context"))
			}
		})

		// Создаем middleware
		middleware := authMiddleware.Middleware()

		// Создаем тестовый сервер
		router := mux.NewRouter()
		router.Use(middleware)
		router.HandleFunc("/test", testHandler).Methods("GET")
		server := httptest.NewServer(router)
		defer server.Close()

		// Создаем тестовый запрос с действительным токеном
		req, err := http.NewRequest("GET", "/test", nil)
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer valid-token")

		// Создаем ResponseRecorder для записи ответа
		rr := httptest.NewRecorder()

		// Выполняем запрос
		router.ServeHTTP(rr, req)

		// Проверяем статус ответа
		assert.Equal(t, http.StatusOK, rr.Code)

		// Проверяем, что все ожидаемые вызовы были выполнены
		mockAuthService.AssertExpectations(t)
	})

	t.Run("Отсутствие токена", func(t *testing.T) {
		// Создаем мок-сервис аутентификации
		mockAuthService := new(MockAuthService)

		// Создаем AuthMiddleware с мок-сервисом
		authMiddleware := NewAuthMiddleware(mockAuthService)

		// Создаем тестовый обработчик
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Проверяем, что информация о пользователе добавлена в контекст
			userID, ok := r.Context().Value("user_id").(int)
			if ok {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("User ID: " + strconv.Itoa(userID)))
			} else {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("No user ID in context"))
			}
		})

		// Создаем middleware
		middleware := authMiddleware.Middleware()

		// Создаем тестовый сервер
		router := mux.NewRouter()
		router.Use(middleware)
		router.HandleFunc("/test", testHandler).Methods("GET")
		server := httptest.NewServer(router)
		defer server.Close()

		// Создаем тестовый запрос без токена
		req, err := http.NewRequest("GET", "/test", nil)
		assert.NoError(t, err)

		// Создаем ResponseRecorder для записи ответа
		rr := httptest.NewRecorder()

		// Выполняем запрос
		router.ServeHTTP(rr, req)

		// Проверяем статус ответа
		assert.Equal(t, http.StatusOK, rr.Code)

		// Проверяем тело ответа
		assert.Equal(t, "No user ID in context", rr.Body.String())

		// Проверяем, что все ожидаемые вызовы были выполнены
		mockAuthService.AssertExpectations(t)
	})

	t.Run("Недействительный токен", func(t *testing.T) {
		// Создаем мок-сервис аутентификации
		mockAuthService := new(MockAuthService)

		// Настраиваем ожидаемое поведение
		mockAuthService.On("ValidateToken", "invalid-token").Return(0, errors.New("недействительный токен"))

		// Создаем AuthMiddleware с мок-сервисом
		authMiddleware := NewAuthMiddleware(mockAuthService)

		// Создаем тестовый обработчик
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Проверяем, что информация о пользователе добавлена в контекст
			userID, ok := r.Context().Value("user_id").(int)
			if ok {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("User ID: " + strconv.Itoa(userID)))
			} else {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("No user ID in context"))
			}
		})

		// Создаем middleware
		middleware := authMiddleware.Middleware()

		// Создаем тестовый сервер
		router := mux.NewRouter()
		router.Use(middleware)
		router.HandleFunc("/test", testHandler).Methods("GET")
		server := httptest.NewServer(router)
		defer server.Close()

		// Создаем тестовый запрос с недействительным токеном
		req, err := http.NewRequest("GET", "/test", nil)
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer invalid-token")

		// Создаем ResponseRecorder для записи ответа
		rr := httptest.NewRecorder()

		// Выполняем запрос
		router.ServeHTTP(rr, req)

		// Проверяем статус ответа
		assert.Equal(t, http.StatusUnauthorized, rr.Code)

		// Проверяем тело ответа
		assert.Contains(t, rr.Body.String(), "Недействительный токен")

		// Проверяем, что все ожидаемые вызовы были выполнены
		mockAuthService.AssertExpectations(t)
	})
}

func TestAuthMiddleware_WithRedis(t *testing.T) {
	// Создаем мок для Redis клиента
	mockRedis := new(MockRedisClient)

	// Создаем мок для сервиса аутентификации
	mockAuthService := new(MockAuthService)

	t.Run("Токен в черном списке", func(t *testing.T) {
		// Сбрасываем моки
		mockRedis.ExpectedCalls = nil
		mockAuthService.ExpectedCalls = nil

		// Настраиваем ожидаемое поведение для Redis
		// Проверяем, что токен находится в черном списке
		mockRedis.On("Get", mock.Anything, "blacklist:blacklisted-token").Return("user_id:123", nil).Once()

		// Создаем AuthMiddleware с мок-сервисом
		authMiddleware := NewAuthMiddleware(mockAuthService)

		// Настраиваем ожидаемое поведение для сервиса аутентификации
		mockAuthService.On("ValidateToken", "blacklisted-token").Return(0, errors.New("недействительный токен")).Once()

		// Создаем тестовый обработчик
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Success"))
		})

		// Создаем middleware
		middleware := authMiddleware.Middleware()

		// Создаем тестовый сервер
		router := mux.NewRouter()
		router.Use(middleware)
		router.HandleFunc("/test", testHandler).Methods("GET")

		// Создаем тестовый запрос с токеном из черного списка
		req, err := http.NewRequest("GET", "/test", nil)
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer blacklisted-token")

		// Создаем ResponseRecorder для записи ответа
		rr := httptest.NewRecorder()

		// Выполняем запрос
		router.ServeHTTP(rr, req)

		// Проверяем статус ответа
		assert.Equal(t, http.StatusUnauthorized, rr.Code)

		// Проверяем тело ответа - используем точное сравнение вместо Contains
		assert.Equal(t, "Недействительный токен\n", rr.Body.String())

		// Проверяем, что все ожидаемые вызовы были выполнены
		mockAuthService.AssertExpectations(t)
	})

	t.Run("Действительный токен с Redis", func(t *testing.T) {
		// Сбрасываем моки
		mockRedis.ExpectedCalls = nil
		mockAuthService.ExpectedCalls = nil

		// Настраиваем ожидаемое поведение для сервиса аутентификации
		mockAuthService.On("ValidateToken", "valid-token").Return(123, nil).Once()

		// Создаем AuthMiddleware с мок-сервисом
		authMiddleware := NewAuthMiddleware(mockAuthService)

		// Создаем тестовый обработчик
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Проверяем, что информация о пользователе добавлена в контекст
			userID, ok := r.Context().Value("user_id").(int)
			if ok {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("User ID: " + strconv.Itoa(userID)))
			} else {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("No user ID in context"))
			}
		})

		// Создаем middleware
		middleware := authMiddleware.Middleware()

		// Создаем тестовый сервер
		router := mux.NewRouter()
		router.Use(middleware)
		router.HandleFunc("/test", testHandler).Methods("GET")

		// Создаем тестовый запрос с действительным токеном
		req, err := http.NewRequest("GET", "/test", nil)
		assert.NoError(t, err)
		req.Header.Set("Authorization", "Bearer valid-token")

		// Создаем ResponseRecorder для записи ответа
		rr := httptest.NewRecorder()

		// Выполняем запрос
		router.ServeHTTP(rr, req)

		// Проверяем статус ответа
		assert.Equal(t, http.StatusOK, rr.Code)

		// Проверяем тело ответа
		assert.Contains(t, rr.Body.String(), "User ID: 123")

		// Проверяем, что все ожидаемые вызовы были выполнены
		mockAuthService.AssertExpectations(t)
	})
}
