package auth

import (
	"database/sql"

	"testing"

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
				mock.ExpectQuery("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)").
					WithArgs("test@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
				mock.ExpectExec("INSERT INTO users").
					WithArgs("test@example.com", "password123", sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 1))
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
				mock.ExpectQuery("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)").
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
				mock.ExpectQuery("SELECT id, email, password, created_at, updated_at FROM users WHERE email = $1").
					WithArgs("test@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password"}).AddRow(1, "test@example.com", "password123"))
			},
			wantErr: false,
		},
		{
			email:    "nonexistent@example.com",
			password: "password123",
			mock: func() {
				mock.ExpectQuery("SELECT id, email, password, created_at, updated_at FROM users WHERE email = $1").
					WithArgs("nonexistent@example.com").
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
		{
			email:    "test@example.com",
			password: "wrongpassword",
			mock: func() {
				mock.ExpectQuery("SELECT id, email, password, created_at, updated_at FROM users WHERE email = $1").
					WithArgs("test@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password"}).AddRow(1, "test@example.com", "password123"))
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
