package middleware

import (
	"delivery/internal/auth"
	"delivery/internal/cache"
	"errors"
	"net/http"
	"net/http/httptest"
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

func TestAuthMiddleware_Middleware(t *testing.T) {
	// Создаем мок для AuthService
	mockAuthService := new(MockAuthService)

	// Создаем AuthMiddleware с мок-сервисом
	authMiddleware := NewAuthMiddleware(mockAuthService)

	// Создаем тестовый обработчик
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, что информация о пользователе добавлена в контекст
		userID, ok := r.Context().Value("user_id").(int)
		if ok {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("User ID: " + string(userID)))
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

	t.Run("Запрос без токена", func(t *testing.T) {
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
	})

	t.Run("Запрос с неверным форматом токена", func(t *testing.T) {
		// Создаем тестовый запрос с неверным форматом токена
		req, err := http.NewRequest("GET", "/test", nil)
		assert.NoError(t, err)
		req.Header.Set("Authorization", "InvalidFormat")

		// Создаем ResponseRecorder для записи ответа
		rr := httptest.NewRecorder()

		// Выполняем запрос
		router.ServeHTTP(rr, req)

		// Проверяем статус ответа
		assert.Equal(t, http.StatusUnauthorized, rr.Code)

		// Проверяем тело ответа
		assert.Contains(t, rr.Body.String(), "Неверный формат токена")
	})

	t.Run("Запрос с недействительным токеном", func(t *testing.T) {
		// Настраиваем мок для проверки токена
		mockAuthService.On("ValidateToken", "invalid-token").Return(0, errors.New("недействительный токен")).Once()

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

	t.Run("Запрос с действительным токеном", func(t *testing.T) {
		// Настраиваем мок для проверки токена
		mockAuthService.On("ValidateToken", "valid-token").Return(123, nil).Once()

		// Создаем тестовый обработчик, который проверяет контекст
		contextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Проверяем, что информация о пользователе добавлена в контекст
			userID, ok := r.Context().Value("user_id").(int)
			assert.True(t, ok)
			assert.Equal(t, 123, userID)

			role, ok := r.Context().Value("user_role").(string)
			assert.True(t, ok)
			assert.Equal(t, "client", role)

			token, ok := r.Context().Value("token").(string)
			assert.True(t, ok)
			assert.Equal(t, "valid-token", token)

			w.WriteHeader(http.StatusOK)
		})

		// Создаем новый роутер с новым обработчиком
		router := mux.NewRouter()
		router.Use(middleware)
		router.HandleFunc("/test", contextHandler).Methods("GET")

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
}
