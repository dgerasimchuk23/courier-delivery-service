package auth

import "delivery/internal/cache"

// AuthServiceInterface определяет интерфейс для сервиса аутентификации
type AuthServiceInterface interface {
	// RegisterUser регистрирует нового пользователя
	RegisterUser(email, password string) error

	// GenerateTokens генерирует токены для пользователя
	GenerateTokens(userID int) (string, string, error)

	// LoginUser аутентифицирует пользователя и возвращает токены
	LoginUser(email, password string) (string, string, error)

	// RefreshToken обновляет токены по refresh токену
	RefreshToken(refreshToken string) (string, string, error)

	// Logout выполняет выход пользователя
	Logout(accessToken, refreshToken string) error

	// ValidateToken проверяет валидность токена и возвращает ID пользователя
	ValidateToken(tokenString string) (int, error)

	// WithCache добавляет клиент кэширования к сервису
	WithCache(cacheClient *cache.RedisClient) *AuthService

	// Close закрывает ресурсы, используемые сервисом
	Close()
}
