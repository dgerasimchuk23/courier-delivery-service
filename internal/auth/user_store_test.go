package auth

import (
	"database/sql"
	"delivery/internal/business/models"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestCreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open a stub database connection: %s", err)
	}
	defer db.Close()

	userStore := NewUserStore(db)

	tests := []struct {
		user    models.User
		mock    func()
		wantErr bool
	}{
		{
			user: models.User{
				Email:    "test@example.com",
				Password: "password123",
				Role:     "client",
			},
			mock: func() {
				mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM users WHERE email = \$1\)`).
					WithArgs("test@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

				// Используем QueryRow вместо Exec для INSERT
				mock.ExpectQuery(`INSERT INTO users \(email, password, role, created_at, updated_at\) VALUES \(\$1, \$2, \$3, \$4, \$4\) RETURNING id`).
					WithArgs("test@example.com", "password123", "client", sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			wantErr: false,
		},
		{
			user: models.User{
				Email:    "test@example.com",
				Password: "password123",
			},
			mock: func() {
				mock.ExpectQuery(`SELECT EXISTS\(SELECT 1 FROM users WHERE email = \$1\)`).
					WithArgs("test@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

				// Используем QueryRow вместо Exec для INSERT
				mock.ExpectQuery(`INSERT INTO users \(email, password, role, created_at, updated_at\) VALUES \(\$1, \$2, \$3, \$4, \$4\) RETURNING id`).
					WithArgs("test@example.com", "password123", "client", sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
			},
			wantErr: false,
		},
		{
			user: models.User{
				Email:    "test@example.com",
				Password: "password123",
			},
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
		_, err := userStore.CreateUser(tt.user)
		if (err != nil) != tt.wantErr {
			t.Errorf("CreateUser() error = %v, wantErr %v", err, tt.wantErr)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	}
}

func TestGetUserByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open a stub database connection: %s", err)
	}
	defer db.Close()

	userStore := NewUserStore(db)

	// Создаем фиксированное время для тестов
	now := time.Now().UTC()

	tests := []struct {
		email   string
		mock    func()
		wantErr bool
	}{
		{
			email: "test@example.com",
			mock: func() {
				mock.ExpectQuery(`SELECT id, email, password, role, created_at, updated_at FROM users WHERE email = \$1`).
					WithArgs("test@example.com").
					WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password", "role", "created_at", "updated_at"}).
						AddRow(1, "test@example.com", "password123", "client", now, now))
			},
			wantErr: false,
		},
		{
			email: "nonexistent@example.com",
			mock: func() {
				mock.ExpectQuery(`SELECT id, email, password, role, created_at, updated_at FROM users WHERE email = \$1`).
					WithArgs("nonexistent@example.com").
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		if tt.mock != nil {
			tt.mock()
		}
		_, err := userStore.GetUserByEmail(tt.email)
		if (err != nil) != tt.wantErr {
			t.Errorf("GetUserByEmail() error = %v, wantErr %v", err, tt.wantErr)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	}
}

func TestGetUserByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open a stub database connection: %s", err)
	}
	defer db.Close()

	userStore := NewUserStore(db)

	// Создаем фиксированное время для тестов
	now := time.Now().UTC()

	tests := []struct {
		id      int
		mock    func()
		wantErr bool
	}{
		{
			id: 1,
			mock: func() {
				mock.ExpectQuery(`SELECT id, email, password, role, created_at, updated_at FROM users WHERE id = \$1`).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"id", "email", "password", "role", "created_at", "updated_at"}).
						AddRow(1, "test@example.com", "password123", "client", now, now))
			},
			wantErr: false,
		},
		{
			id: 2,
			mock: func() {
				mock.ExpectQuery(`SELECT id, email, password, role, created_at, updated_at FROM users WHERE id = \$1`).
					WithArgs(2).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		if tt.mock != nil {
			tt.mock()
		}
		_, err := userStore.GetUserByID(tt.id)
		if (err != nil) != tt.wantErr {
			t.Errorf("GetUserByID() error = %v, wantErr %v", err, tt.wantErr)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	}
}
