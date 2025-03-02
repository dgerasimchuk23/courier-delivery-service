package auth

import (
	"delivery/internal/business/models"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Предоставляет методы для аутентификации и авторизации
type AuthService struct {
	store *UserStore
}

// Создание нового экземпляра AuthService
func NewAuthService(store *UserStore) *AuthService {
	return &AuthService{store: store}
}

// Регистрация нового пользователя
func (s *AuthService) RegisterUser(email, password string) error {
	// Проверка валидности email и пароля
	if email == "" || password == "" {
		return errors.New("email и пароль не могут быть пустыми")
	}

	// Создание пользователя
	user := models.User{
		Email:    email,
		Password: password, // Пароль хранится в открытом виде
	}

	_, err := s.store.CreateUser(user)
	if err != nil {
		return fmt.Errorf("ошибка при регистрации пользователя: %w", err)
	}

	return nil
}

// Генерация токенов для пользователя
func (s *AuthService) GenerateTokens(userID int) (string, string) {
	// Генерация access токена
	accessToken := fmt.Sprintf("access_token_%d", userID)

	// Генерация refresh токена с использованием UUID
	refreshToken := uuid.NewString()

	return accessToken, refreshToken
}

// Аутентификация пользователя и возвращение токенов
func (s *AuthService) LoginUser(email, password string) (string, string, error) {
	// Получение пользователя по email
	user, err := s.store.GetUserByEmail(email)
	if err != nil {
		return "", "", fmt.Errorf("ошибка при аутентификации: %w", err)
	}

	// Проверка пароля
	if user.Password != password {
		return "", "", errors.New("неверный пароль")
	}

	// Генерация токенов
	accessToken, refreshToken := s.GenerateTokens(user.ID)

	// Сохранение refresh токена в БД
	tokenExpiry := time.Now().UTC().Add(30 * 24 * time.Hour) // 30 дней
	tokenModel := models.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: tokenExpiry,
	}

	err = s.store.SaveRefreshToken(tokenModel)
	if err != nil {
		return "", "", fmt.Errorf("ошибка при сохранении токена: %w", err)
	}

	return accessToken, refreshToken, nil
}
