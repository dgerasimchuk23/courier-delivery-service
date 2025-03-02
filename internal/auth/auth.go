package auth

import "delivery/internal/business/models"

// Интерфейс для аутентификации пользователей
type Authenticator interface {
	// Регистрация нового пользователя
	RegisterUser(email, password string) error

	// Аутентификация пользователя и возвращение токенов
	LoginUser(email, password string) (accessToken, refreshToken string, err error)

	// Обновление токенов по refresh токену
	RefreshToken(refreshToken string) (newAccessToken, newRefreshToken string, err error)

	// Выход пользователя (логаут)
	Logout(accessToken, refreshToken string) error
}

// Интерфейс для работы с пользователями
type UserRepository interface {
	// Создание нового пользователя
	CreateUser(user models.User) (int, error)

	// Возвращение пользователя по email
	GetUserByEmail(email string) (models.User, error)

	// Возвращение пользователя по ID
	GetUserByID(id int) (models.User, error)

	// Сохранение refresh токена
	SaveRefreshToken(token models.RefreshToken) error

	// Возвращение refresh токена по значению
	GetRefreshToken(tokenString string) (models.RefreshToken, error)

	// Удаление refresh токена
	DeleteRefreshToken(tokenString string) error

	// Удаление всех истекших токенов
	DeleteExpiredTokens() error
}
