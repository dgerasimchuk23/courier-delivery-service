package auth

import (
	"database/sql"
	"delivery/internal/business/models"
	"errors"
	"fmt"
	"time"
)

var (
	ErrUserNotFound       = errors.New("пользователь не найден")
	ErrEmailAlreadyExists = errors.New("пользователь с таким email уже существует")
)

// UserStore представляет хранилище пользователей
type UserStore struct {
	db *sql.DB
}

// NewUserStore создает новый экземпляр UserStore
func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

// CreateUser создает нового пользователя
func (s *UserStore) CreateUser(user models.User) (int, error) {
	// Проверяем, существует ли пользователь с таким email
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", user.Email).Scan(&exists)
	if err != nil {
		return 0, fmt.Errorf("ошибка при проверке существования пользователя: %w", err)
	}
	if exists {
		return 0, errors.New("пользователь с таким email уже существует")
	}

	// Если роль не указана, устанавливаем значение по умолчанию
	if user.Role == "" {
		user.Role = "client"
	}

	// Создаем пользователя
	var id int
	err = s.db.QueryRow(
		"INSERT INTO users (email, password, role, created_at, updated_at) VALUES ($1, $2, $3, $4, $4) RETURNING id",
		user.Email, user.Password, user.Role, time.Now(),
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("ошибка при создании пользователя: %w", err)
	}

	return id, nil
}

// GetUserByEmail получает пользователя по email
func (s *UserStore) GetUserByEmail(email string) (models.User, error) {
	var user models.User
	err := s.db.QueryRow(
		"SELECT id, email, password, role, created_at, updated_at FROM users WHERE email = $1",
		email,
	).Scan(&user.ID, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, errors.New("пользователь не найден")
		}
		return models.User{}, fmt.Errorf("ошибка при получении пользователя: %w", err)
	}

	return user, nil
}

// GetUserByID получает пользователя по ID
func (s *UserStore) GetUserByID(id int) (models.User, error) {
	var user models.User
	err := s.db.QueryRow(
		"SELECT id, email, password, role, created_at, updated_at FROM users WHERE id = $1",
		id,
	).Scan(&user.ID, &user.Email, &user.Password, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, errors.New("пользователь не найден")
		}
		return models.User{}, fmt.Errorf("ошибка при получении пользователя: %w", err)
	}

	return user, nil
}

// SaveRefreshToken сохраняет refresh токен в базе данных
func (s *UserStore) SaveRefreshToken(token models.RefreshToken) error {
	_, err := s.db.Exec(
		"INSERT INTO refresh_tokens (user_id, token, expires_at, created_at) VALUES ($1, $2, $3, $4)",
		token.UserID, token.Token, token.ExpiresAt, time.Now(),
	)
	if err != nil {
		return fmt.Errorf("ошибка при сохранении refresh токена: %w", err)
	}

	return nil
}

// GetRefreshToken получает refresh токен из базы данных
func (s *UserStore) GetRefreshToken(token string) (models.RefreshToken, error) {
	var refreshToken models.RefreshToken
	err := s.db.QueryRow(
		"SELECT user_id, token, expires_at, created_at FROM refresh_tokens WHERE token = $1",
		token,
	).Scan(&refreshToken.UserID, &refreshToken.Token, &refreshToken.ExpiresAt, &refreshToken.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.RefreshToken{}, errors.New("токен не найден")
		}
		return models.RefreshToken{}, fmt.Errorf("ошибка при получении refresh токена: %w", err)
	}

	return refreshToken, nil
}

// DeleteRefreshToken удаляет refresh токен из базы данных
func (s *UserStore) DeleteRefreshToken(token string) error {
	_, err := s.db.Exec("DELETE FROM refresh_tokens WHERE token = $1", token)
	if err != nil {
		return fmt.Errorf("ошибка при удалении refresh токена: %w", err)
	}

	return nil
}

// DeleteExpiredRefreshTokens удаляет просроченные refresh токены
func (s *UserStore) DeleteExpiredRefreshTokens() error {
	_, err := s.db.Exec("DELETE FROM refresh_tokens WHERE expires_at < $1", time.Now())
	if err != nil {
		return fmt.Errorf("ошибка при удалении просроченных refresh токенов: %w", err)
	}

	return nil
}
