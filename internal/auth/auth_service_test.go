package auth

import (
	"database/sql"
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
