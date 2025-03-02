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

// Предоставляет методы для работы с пользователями в БД
type UserStore struct {
	db *sql.DB
}

// Создание нового экземпляра UserStore
func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

// Создание нового пользователя в БД
func (s *UserStore) CreateUser(user models.User) (int, error) {
	// Проверка существования пользователя с таким email
	var exists bool
	err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", user.Email).Scan(&exists)
	if err != nil {
		return 0, fmt.Errorf("ошибка при проверке существования пользователя: %w", err)
	}

	if exists {
		return 0, ErrEmailAlreadyExists
	}

	// Создание пользователя
	var id int
	now := time.Now().UTC()
	err = s.db.QueryRow(
		"INSERT INTO users (email, password, created_at, updated_at) VALUES ($1, $2, $3, $4) RETURNING id",
		user.Email, user.Password, now, now,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("ошибка при создании пользователя: %w", err)
	}

	return id, nil
}

// Возвращение пользователя по email
func (s *UserStore) GetUserByEmail(email string) (models.User, error) {
	var user models.User
	err := s.db.QueryRow(
		"SELECT id, email, password, created_at, updated_at FROM users WHERE email = $1",
		email,
	).Scan(&user.ID, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, fmt.Errorf("ошибка при получении пользователя: %w", err)
	}

	return user, nil
}

// Возвращение пользователя по ID
func (s *UserStore) GetUserByID(id int) (models.User, error) {
	var user models.User
	err := s.db.QueryRow(
		"SELECT id, email, password, created_at, updated_at FROM users WHERE id = $1",
		id,
	).Scan(&user.ID, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, fmt.Errorf("ошибка при получении пользователя: %w", err)
	}

	return user, nil
}

// Сохранение refresh токена в БД
func (s *UserStore) SaveRefreshToken(token models.RefreshToken) error {
	_, err := s.db.Exec(
		"INSERT INTO refresh_tokens (user_id, token, expires_at, created_at) VALUES ($1, $2, $3, $4)",
		token.UserID, token.Token, token.ExpiresAt, time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("ошибка при сохранении refresh токена: %w", err)
	}
	return nil
}

// Возвращение refresh токена по значению токена
func (s *UserStore) GetRefreshToken(tokenString string) (models.RefreshToken, error) {
	var token models.RefreshToken
	err := s.db.QueryRow(
		"SELECT id, user_id, token, expires_at, created_at FROM refresh_tokens WHERE token = $1",
		tokenString,
	).Scan(&token.ID, &token.UserID, &token.Token, &token.ExpiresAt, &token.CreatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.RefreshToken{}, errors.New("токен не найден")
		}
		return models.RefreshToken{}, fmt.Errorf("ошибка при получении токена: %w", err)
	}

	return token, nil
}

// Удаление refresh токена из БД
func (s *UserStore) DeleteRefreshToken(tokenString string) error {
	_, err := s.db.Exec("DELETE FROM refresh_tokens WHERE token = $1", tokenString)
	if err != nil {
		return fmt.Errorf("ошибка при удалении токена: %w", err)
	}
	return nil
}

// Удаление всех истекших токенов
func (s *UserStore) DeleteExpiredTokens() error {
	_, err := s.db.Exec("DELETE FROM refresh_tokens WHERE expires_at < $1", time.Now().UTC())
	if err != nil {
		return fmt.Errorf("ошибка при удалении истекших токенов: %w", err)
	}
	return nil
}
