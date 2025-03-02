package auth

import (
	"context"
	"delivery/internal/business/models"
	"delivery/internal/cache"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Предоставляет методы для аутентификации и авторизации
type AuthService struct {
	store       *UserStore
	cacheClient *cache.RedisClient
}

// Создание нового экземпляра AuthService
func NewAuthService(store *UserStore) *AuthService {
	return &AuthService{store: store}
}

// WithCache добавляет клиент кэширования к сервису
func (s *AuthService) WithCache(cacheClient *cache.RedisClient) *AuthService {
	s.cacheClient = cacheClient
	return s
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

// Обновление токенов по refresh токену
func (s *AuthService) RefreshToken(refreshToken string) (string, string, error) {
	// Получаем refresh токен из БД
	token, err := s.store.GetRefreshToken(refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("ошибка при получении refresh токена: %w", err)
	}

	// Проверяем, не истек ли токен
	if token.ExpiresAt.Before(time.Now().UTC()) {
		return "", "", errors.New("refresh токен истек")
	}

	// Удаляем старый refresh токен
	err = s.store.DeleteRefreshToken(refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("ошибка при удалении старого refresh токена: %w", err)
	}

	// Генерируем новые токены
	accessToken, newRefreshToken := s.GenerateTokens(token.UserID)

	// Сохраняем новый refresh токен
	tokenExpiry := time.Now().UTC().Add(30 * 24 * time.Hour) // 30 дней
	newTokenModel := models.RefreshToken{
		UserID:    token.UserID,
		Token:     newRefreshToken,
		ExpiresAt: tokenExpiry,
	}

	err = s.store.SaveRefreshToken(newTokenModel)
	if err != nil {
		return "", "", fmt.Errorf("ошибка при сохранении нового refresh токена: %w", err)
	}

	return accessToken, newRefreshToken, nil
}

// Выход пользователя (логаут)
func (s *AuthService) Logout(accessToken, refreshToken string) error {
	// Удаляем refresh токен из БД
	err := s.store.DeleteRefreshToken(refreshToken)
	if err != nil {
		return fmt.Errorf("ошибка при удалении refresh токена: %w", err)
	}

	// Добавляем access токен в черный список, если доступен Redis
	if s.cacheClient != nil {
		ctx := context.Background()

		// Извлекаем ID пользователя из токена для определения времени жизни
		var userID int
		_, scanErr := fmt.Sscanf(accessToken, "access_token_%d", &userID)
		if scanErr != nil {
			return fmt.Errorf("ошибка при извлечении ID пользователя из токена: %w", scanErr)
		}

		// В реальном приложении здесь нужно извлечь время истечения из JWT
		// Сейчас просто устанавливаем TTL в 7 минут (как указано в требованиях)
		expiration := 7 * time.Minute

		blacklistKey := fmt.Sprintf("blacklist:%s", accessToken)
		err := s.cacheClient.Set(ctx, blacklistKey, "revoked", expiration)
		if err != nil {
			return fmt.Errorf("ошибка при добавлении токена в черный список: %w", err)
		}
	}

	return nil
}
