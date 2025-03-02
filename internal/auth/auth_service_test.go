package auth

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestRegisterUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open a stub database connection: %s", err)
	}
	defer db.Close()

	userStore := NewUserStore(db)
	authService := NewAuthService(userStore)

	tests := []struct {
		email    string
		password string
		mock     func()
		wantErr  bool
	}{
		{
			email:    "test@example.com",
			password: "password123",
			mock: func() {
				mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM users WHERE email = \$1\)`).
					WithArgs("test@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

				// Используем QueryRow вместо Exec для INSERT
				mock.ExpectQuery(`INSERT INTO users \(email, password, created_at, updated_at\) VALUES \(\$1, \$2, \$3, \$4\) RETURNING id`).
					WithArgs("test@example.com", "password123", sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			wantErr: false,
		},
		{
			email:    "",
			password: "password123",
			wantErr:  true,
		},
		{
			email:    "test@example.com",
			password: "password123",
			mock: func() {
				mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM users WHERE email = \$1\)`).
					WithArgs("test@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		if tt.mock != nil {
			tt.mock()
		}
		err := authService.RegisterUser(tt.email, tt.password)
		if (err != nil) != tt.wantErr {
			t.Errorf("RegisterUser() error = %v, wantErr %v", err, tt.wantErr)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	}
}

func TestLoginUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open a stub database connection: %s", err)
	}
	defer db.Close()

	userStore := NewUserStore(db)
	authService := NewAuthService(userStore)

	// Создаем фиксированное время для тестов
	now := time.Now().UTC()

	tests := []struct {
		email    string
		password string
		mock     func()
		wantErr  bool
	}{
		{
			email:    "test@example.com",
			password: "password123",
			mock: func() {
				mock.ExpectQuery(`SELECT id, email, password, created_at, updated_at FROM users WHERE email = \$1`).
					WithArgs("test@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password", "created_at", "updated_at"}).
						AddRow(1, "test@example.com", "password123", now, now))

				// Добавляем ожидание для GenerateTokens
				mock.ExpectExec(`INSERT INTO refresh_tokens \(user_id, token, expires_at, created_at\) VALUES \(\$1, \$2, \$3, \$4\)`).
					WithArgs(1, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			email:    "nonexistent@example.com",
			password: "password123",
			mock: func() {
				mock.ExpectQuery(`SELECT id, email, password, created_at, updated_at FROM users WHERE email = \$1`).
					WithArgs("nonexistent@example.com").
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
		{
			email:    "test@example.com",
			password: "wrongpassword",
			mock: func() {
				mock.ExpectQuery(`SELECT id, email, password, created_at, updated_at FROM users WHERE email = \$1`).
					WithArgs("test@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password", "created_at", "updated_at"}).
						AddRow(1, "test@example.com", "password123", now, now))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		if tt.mock != nil {
			tt.mock()
		}
		_, _, err := authService.LoginUser(tt.email, tt.password)
		if (err != nil) != tt.wantErr {
			t.Errorf("LoginUser() error = %v, wantErr %v", err, tt.wantErr)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	}
}

func TestRefreshToken(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open a stub database connection: %s", err)
	}
	defer db.Close()

	userStore := NewUserStore(db)
	authService := NewAuthService(userStore)

	// Создаем фиксированное время для тестов
	now := time.Now().UTC()
	futureTime := now.Add(24 * time.Hour) // Токен действителен
	pastTime := now.Add(-24 * time.Hour)  // Токен истек

	tests := []struct {
		name         string
		refreshToken string
		mock         func()
		wantErr      bool
	}{
		{
			name:         "Valid refresh token",
			refreshToken: "valid-refresh-token",
			mock: func() {
				// Получение refresh токена
				mock.ExpectQuery(`SELECT id, user_id, token, expires_at, created_at FROM refresh_tokens WHERE token = \$1`).
					WithArgs("valid-refresh-token").
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "token", "expires_at", "created_at"}).
						AddRow(1, 1, "valid-refresh-token", futureTime, now))

				// Удаление старого токена
				mock.ExpectExec(`DELETE FROM refresh_tokens WHERE token = \$1`).
					WithArgs("valid-refresh-token").
					WillReturnResult(sqlmock.NewResult(0, 1))

				// Сохранение нового токена
				mock.ExpectExec(`INSERT INTO refresh_tokens \(user_id, token, expires_at, created_at\) VALUES \(\$1, \$2, \$3, \$4\)`).
					WithArgs(1, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name:         "Expired refresh token",
			refreshToken: "expired-refresh-token",
			mock: func() {
				mock.ExpectQuery(`SELECT id, user_id, token, expires_at, created_at FROM refresh_tokens WHERE token = \$1`).
					WithArgs("expired-refresh-token").
					WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "token", "expires_at", "created_at"}).
						AddRow(2, 1, "expired-refresh-token", pastTime, now))
			},
			wantErr: true,
		},
		{
			name:         "Non-existent refresh token",
			refreshToken: "non-existent-token",
			mock: func() {
				mock.ExpectQuery(`SELECT id, user_id, token, expires_at, created_at FROM refresh_tokens WHERE token = \$1`).
					WithArgs("non-existent-token").
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mock != nil {
				tt.mock()
			}

			accessToken, refreshToken, err := authService.RefreshToken(tt.refreshToken)
			if (err != nil) != tt.wantErr {
				t.Errorf("RefreshToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if accessToken == "" {
					t.Error("RefreshToken() accessToken is empty")
				}
				if refreshToken == "" {
					t.Error("RefreshToken() refreshToken is empty")
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestLogout(t *testing.T) {
	// Пропускаем тест, если нет возможности создать мок Redis
	t.Skip("Skipping test that requires Redis mock")

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open a stub database connection: %s", err)
	}
	defer db.Close()

	userStore := NewUserStore(db)
	authService := NewAuthService(userStore)

	// В реальном тесте здесь должен быть настоящий Redis клиент или его мок
	// Сейчас просто проверяем базовую функциональность без Redis

	tests := []struct {
		name         string
		accessToken  string
		refreshToken string
		mock         func()
		wantErr      bool
	}{
		{
			name:         "Successful logout",
			accessToken:  "access_token_1",
			refreshToken: "valid-refresh-token",
			mock: func() {
				mock.ExpectExec(`DELETE FROM refresh_tokens WHERE token = \$1`).
					WithArgs("valid-refresh-token").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:         "Error deleting refresh token",
			accessToken:  "access_token_1",
			refreshToken: "error-token",
			mock: func() {
				mock.ExpectExec(`DELETE FROM refresh_tokens WHERE token = \$1`).
					WithArgs("error-token").
					WillReturnError(errors.New("database error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mock != nil {
				tt.mock()
			}

			err := authService.Logout(tt.accessToken, tt.refreshToken)
			if (err != nil) != tt.wantErr {
				t.Errorf("Logout() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
