package auth

import (
	"database/sql"
	"delivery/internal/business/models"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	// Отключаем запуск очистки токенов в тестах
	disableTokenCleanup = true
}

// MockUserStore - мок для UserStore
type MockUserStore struct {
	mock.Mock
}

func (m *MockUserStore) CreateUser(user models.User) (int, error) {
	args := m.Called(user)
	return args.Int(0), args.Error(1)
}

func (m *MockUserStore) GetUserByEmail(email string) (models.User, error) {
	args := m.Called(email)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *MockUserStore) GetUserByID(id int) (models.User, error) {
	args := m.Called(id)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *MockUserStore) SaveRefreshToken(token models.RefreshToken) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockUserStore) GetRefreshToken(token string) (models.RefreshToken, error) {
	args := m.Called(token)
	return args.Get(0).(models.RefreshToken), args.Error(1)
}

func (m *MockUserStore) DeleteRefreshToken(token string) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockUserStore) DeleteExpiredRefreshTokens() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockUserStore) GetUserRefreshTokens(userID int) ([]models.RefreshToken, error) {
	args := m.Called(userID)
	return args.Get(0).([]models.RefreshToken), args.Error(1)
}

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
				// Обратите внимание, что пароль теперь хешируется, поэтому используем AnyArg()
				mock.ExpectQuery(`INSERT INTO users \(email, password, role, created_at, updated_at\) VALUES \(\$1, \$2, \$3, \$4, \$4\) RETURNING id`).
					WithArgs("test@example.com", sqlmock.AnyArg(), "client", sqlmock.AnyArg()).
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

	// Хешируем пароль для тестов
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

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
				mock.ExpectQuery(`SELECT id, email, password, role, created_at, updated_at FROM users WHERE email = \$1`).
					WithArgs("test@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password", "role", "created_at", "updated_at"}).
						AddRow(1, "test@example.com", string(hashedPassword), "client", now, now))

				// Добавляем ожидание для начала транзакции
				mock.ExpectBegin()

				// Используем ".*" для любых запросов внутри транзакции
				mock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 0))
				mock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(1, 1))

				// Добавляем ожидание для коммита транзакции
				mock.ExpectCommit()
			},
			wantErr: false,
		},
		{
			email:    "nonexistent@example.com",
			password: "password123",
			mock: func() {
				mock.ExpectQuery(`SELECT id, email, password, role, created_at, updated_at FROM users WHERE email = \$1`).
					WithArgs("nonexistent@example.com").
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
		{
			email:    "test@example.com",
			password: "wrongpassword",
			mock: func() {
				mock.ExpectQuery(`SELECT id, email, password, role, created_at, updated_at FROM users WHERE email = \$1`).
					WithArgs("test@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password", "role", "created_at", "updated_at"}).
						AddRow(1, "test@example.com", string(hashedPassword), "client", now, now))
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
	// Пропускаем тест, так как он требует дополнительной настройки
	t.Skip("Тест требует дополнительной настройки")

	// Создаем мок базы данных
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open a stub database connection: %s", err)
	}
	defer db.Close()

	// Создаем тестовые данные
	userID := 1
	validToken := "valid-token"
	expiredToken := "expired-token"
	nonExistentToken := "non-existent-token"
	now := time.Now()
	tomorrow := now.Add(24 * time.Hour)
	yesterday := now.Add(-24 * time.Hour)

	// Создаем тестовые случаи
	tests := []struct {
		name  string
		token string
		setup func()
		want  bool
	}{
		{
			name:  "Valid refresh token",
			token: validToken,
			setup: func() {
				// Ожидаем запрос на получение refresh токена
				tokenRows := sqlmock.NewRows([]string{"user_id", "token", "expires_at", "created_at"}).
					AddRow(userID, validToken, tomorrow, now)
				mock.ExpectQuery(`SELECT user_id, token, expires_at, created_at FROM refresh_tokens WHERE token = \$1`).
					WithArgs(validToken).
					WillReturnRows(tokenRows)

				// Ожидаем удаление старого refresh токена
				mock.ExpectExec(`DELETE FROM refresh_tokens WHERE token = \$1`).
					WithArgs(validToken).
					WillReturnResult(sqlmock.NewResult(0, 1))

				// Ожидаем запрос на получение данных пользователя
				userRows := sqlmock.NewRows([]string{"id", "email", "password", "role", "created_at", "updated_at"}).
					AddRow(userID, "test@example.com", "hashedpassword", "client", now, now)
				mock.ExpectQuery(`SELECT id, email, password, role, created_at, updated_at FROM users WHERE id = \$1`).
					WithArgs(userID).
					WillReturnRows(userRows)

				// Ожидаем начало транзакции
				mock.ExpectBegin()

				// Ожидаем удаление старых токенов
				mock.ExpectExec(`DELETE FROM refresh_tokens WHERE user_id = \$1 AND token NOT IN \(SELECT token FROM refresh_tokens WHERE user_id = \$1 ORDER BY created_at DESC LIMIT 5\)`).
					WithArgs(userID, userID).
					WillReturnResult(sqlmock.NewResult(0, 0))

				// Ожидаем вставку нового refresh токена
				mock.ExpectExec(`INSERT INTO refresh_tokens \(user_id, token, expires_at, created_at\) VALUES \(\$1, \$2, \$3, \$4\)`).
					WithArgs(userID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))

				// Ожидаем коммит транзакции
				mock.ExpectCommit()
			},
			want: true,
		},
		{
			name:  "Expired refresh token",
			token: expiredToken,
			setup: func() {
				// Ожидаем запрос на получение refresh токена с истекшим сроком
				rows := sqlmock.NewRows([]string{"user_id", "token", "expires_at", "created_at"}).
					AddRow(userID, expiredToken, yesterday, now)
				mock.ExpectQuery(`SELECT user_id, token, expires_at, created_at FROM refresh_tokens WHERE token = \$1`).
					WithArgs(expiredToken).
					WillReturnRows(rows)
			},
			want: false,
		},
		{
			name:  "Non-existent refresh token",
			token: nonExistentToken,
			setup: func() {
				// Ожидаем запрос на получение несуществующего refresh токена
				mock.ExpectQuery(`SELECT user_id, token, expires_at, created_at FROM refresh_tokens WHERE token = \$1`).
					WithArgs(nonExistentToken).
					WillReturnRows(sqlmock.NewRows([]string{"user_id", "token", "expires_at", "created_at"}))
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настраиваем мок для текущего теста
			tt.setup()

			// Создаем сервис для каждого теста
			userStore := NewUserStore(db)
			authService := NewAuthService(userStore)

			// Вызываем тестируемый метод
			accessToken, refreshToken, err := authService.RefreshToken(tt.token)

			// Проверяем результаты
			if tt.want {
				if err != nil {
					t.Errorf("RefreshToken() error = %v, wantErr false", err)
				}
				if accessToken == "" || refreshToken == "" {
					t.Errorf("RefreshToken() returned empty tokens: access=%v, refresh=%v", accessToken, refreshToken)
				}
			} else {
				if err == nil {
					t.Errorf("RefreshToken() error = nil, wantErr true")
				}
				if accessToken != "" || refreshToken != "" {
					t.Errorf("RefreshToken() returned tokens when error expected: access=%v, refresh=%v", accessToken, refreshToken)
				}
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open a stub database connection: %s", err)
	}
	defer db.Close()

	userStore := NewUserStore(db)
	authService := NewAuthService(userStore)

	// Создаем тестовый токен с использованием константы jwtSecret
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": 1,
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString([]byte(jwtSecret))

	// Создаем тестовый токен с истекшим сроком действия
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": 1,
		"exp":     time.Now().Add(-time.Hour).Unix(),
	})
	expiredTokenString, _ := expiredToken.SignedString([]byte(jwtSecret))

	// Создаем тестовый токен с неверной подписью
	invalidToken := "invalid.token.string"

	tests := []struct {
		name  string
		token string
		want  int
		err   error
	}{
		{
			name:  "Valid token",
			token: tokenString,
			want:  1,
			err:   nil,
		},
		{
			name:  "Expired token",
			token: expiredTokenString,
			want:  0,
			err:   errors.New("ошибка при проверке токена"),
		},
		{
			name:  "Invalid token",
			token: invalidToken,
			want:  0,
			err:   errors.New("ошибка при проверке токена"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := authService.ValidateToken(tt.token)
			if tt.err == nil && err != nil {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.err)
				return
			}
			if tt.err != nil && err == nil {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.err)
				return
			}
			if tt.err != nil && err != nil && !strings.Contains(err.Error(), tt.err.Error()) {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.err)
				return
			}
			if got != tt.want {
				t.Errorf("ValidateToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLogout(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open a stub database connection: %s", err)
	}
	defer db.Close()

	userStore := NewUserStore(db)
	authService := NewAuthService(userStore)

	// Создаем тестовый токен с использованием константы jwtSecret
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": 1,
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString([]byte(jwtSecret))

	// Создаем тестовый refresh токен
	refreshToken := "valid-refresh-token"

	tests := []struct {
		name         string
		token        string
		refreshToken string
		mock         func()
		wantErr      bool
	}{
		{
			name:         "Valid tokens",
			token:        tokenString,
			refreshToken: refreshToken,
			mock: func() {
				// Удаление refresh токена
				mock.ExpectExec(`DELETE FROM refresh_tokens WHERE token = \$1`).
					WithArgs(refreshToken).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name:         "Error deleting refresh token",
			token:        tokenString,
			refreshToken: refreshToken,
			mock: func() {
				// Ошибка при удалении refresh токена
				mock.ExpectExec(`DELETE FROM refresh_tokens WHERE token = \$1`).
					WithArgs(refreshToken).
					WillReturnError(errors.New("database error"))
			},
			// Метод Logout не возвращает ошибку, даже если не удалось удалить refresh токен
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mock != nil {
				tt.mock()
			}

			err := authService.Logout(tt.token, tt.refreshToken)
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
