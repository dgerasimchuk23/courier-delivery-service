package auth

import (
	"context"
	"delivery/internal/cache"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// JWTMiddleware представляет middleware для проверки JWT токенов
type JWTMiddleware struct {
	authService *AuthService
	cacheClient *cache.RedisClient
}

// NewJWTMiddleware создает новый экземпляр JWTMiddleware
func NewJWTMiddleware(authService *AuthService, cacheClient *cache.RedisClient) *JWTMiddleware {
	return &JWTMiddleware{
		authService: authService,
		cacheClient: cacheClient,
	}
}

// Middleware возвращает HTTP middleware для проверки JWT токенов
func (m *JWTMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Получаем токен из заголовка Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Отсутствует заголовок Authorization", http.StatusUnauthorized)
			return
		}

		// Проверяем формат токена
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Неверный формат токена", http.StatusUnauthorized)
			return
		}

		token := parts[1]

		// Проверяем, находится ли токен в черном списке
		if m.cacheClient != nil {
			ctx := context.Background()
			blacklistKey := fmt.Sprintf("blacklist:%s", token)
			_, err := m.cacheClient.Get(ctx, blacklistKey)
			if err == nil {
				// Токен найден в черном списке
				http.Error(w, "Токен недействителен", http.StatusUnauthorized)
				return
			}
		}

		// Проверяем валидность токена
		userID, err := m.authService.ValidateToken(token)
		if err != nil {
			http.Error(w, "Недействительный токен: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Добавляем userID в контекст запроса
		ctx := context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// validateToken проверяет валидность токена и возвращает ID пользователя
func (m *JWTMiddleware) validateToken(token string) (int, error) {
	// В реальном приложении здесь должна быть проверка JWT токена
	// Сейчас просто извлекаем ID из нашего простого формата
	if strings.HasPrefix(token, "access_token_") {
		var userID int
		_, err := fmt.Sscanf(token, "access_token_%d", &userID)
		if err != nil {
			return 0, errors.New("неверный формат токена")
		}
		return userID, nil
	}
	return 0, errors.New("неверный формат токена")
}
