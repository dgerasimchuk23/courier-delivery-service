package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWTMiddleware(t *testing.T) {
	// Создаем тестовый AuthService
	userStore := NewUserStore(nil) // Для этого теста БД не нужна
	authService := NewAuthService(userStore)

	// Создаем middleware
	middleware := NewJWTMiddleware(authService, nil)

	// Создаем тестовый обработчик
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, что userID добавлен в контекст
		userID := r.Context().Value("userID")
		if userID == nil {
			t.Error("userID not found in context")
		}
		if userID.(int) != 123 {
			t.Errorf("expected userID 123, got %v", userID)
		}
		w.WriteHeader(http.StatusOK)
	})

	// Создаем тестовый JWT токен
	claims := &Claims{
		UserID: 123,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "delivery-app",
			Subject:   "123",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(jwtSecret))

	// Создаем тестовый запрос с валидным токеном
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+tokenString)

	// Создаем ResponseRecorder для записи ответа
	rr := httptest.NewRecorder()

	// Обрабатываем запрос через middleware
	middleware.Middleware(nextHandler).ServeHTTP(rr, req)

	// Проверяем статус ответа
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Тест с отсутствующим заголовком Authorization
	req, _ = http.NewRequest("GET", "/", nil)
	rr = httptest.NewRecorder()
	middleware.Middleware(nextHandler).ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}

	// Тест с неверным форматом токена
	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	rr = httptest.NewRecorder()
	middleware.Middleware(nextHandler).ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}

	// Тест с недействительным токеном
	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rr = httptest.NewRecorder()
	middleware.Middleware(nextHandler).ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rr.Code)
	}
}
