package middleware

import (
	"context"
	"delivery/internal/auth"
	"delivery/internal/cache"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// Определяем строковые константы для ключей контекста
const (
	UserIDKey   = "user_id"
	UserRoleKey = "user_role"
	TokenKey    = "token"
)

// AuthMiddleware представляет middleware для аутентификации
type AuthMiddleware struct {
	authService auth.AuthServiceInterface
	redisClient *cache.RedisClient
}

// NewAuthMiddleware создает новый экземпляр AuthMiddleware
func NewAuthMiddleware(authService auth.AuthServiceInterface) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// WithRedis добавляет Redis клиент к middleware
func (am *AuthMiddleware) WithRedis(redisClient *cache.RedisClient) *AuthMiddleware {
	am.redisClient = redisClient
	return am
}

// Middleware возвращает middleware для аутентификации
func (am *AuthMiddleware) Middleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Получаем токен из заголовка Authorization
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				// Если токен отсутствует, просто передаем запрос дальше
				// Пользователь будет считаться неаутентифицированным
				next.ServeHTTP(w, r)
				return
			}

			// Проверяем формат токена
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				// Неверный формат токена
				http.Error(w, "Неверный формат токена", http.StatusUnauthorized)
				return
			}

			// Получаем токен
			tokenString := parts[1]

			// Проверяем токен
			userID, err := am.authService.ValidateToken(tokenString)
			if err != nil {
				// Токен недействителен
				http.Error(w, "Недействительный токен", http.StatusUnauthorized)
				return
			}

			// Определяем роль пользователя (в реальном приложении это должно быть получено из базы данных)
			// TODO: Получить роль пользователя из базы данных
			role := "client" // По умолчанию считаем, что пользователь - клиент

			// Добавляем информацию о пользователе в контекст
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			ctx = context.WithValue(ctx, UserRoleKey, role)
			ctx = context.WithValue(ctx, TokenKey, tokenString)

			// Передаем запрос дальше с обновленным контекстом
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
